package postgres

import (
	"context"
	"errors"
	"fmt"
	gqlmodel "github.com/cassini-Inner/inner-src-mgmt-go/graph/model"
	dbmodel "github.com/cassini-Inner/inner-src-mgmt-go/postgres/model"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
)

type JobsRepo struct {
	db *sqlx.DB
}

func NewJobsRepo(db *sqlx.DB) *JobsRepo {
	return &JobsRepo{db: db}
}

func (j *JobsRepo) CreateJob(ctx context.Context, input *gqlmodel.CreateJobInput, user *dbmodel.User) (*dbmodel.Job, error) {
	// create a new transaction for creating a job
	tx, err := j.db.Beginx()
	if err != nil {
		return nil, err
	}

	// insert the information into the job table
	var insertedJobId string
	err = tx.QueryRowContext(ctx, createJobQuery, input.Title, input.Desc, input.Difficulty, user.Id).Scan(&insertedJobId)

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	var insertedMilestoneIds []int
	stmt, valueArgs := getInsertMilestonesStatement(input, insertedJobId, user.Id)
	stmt = j.db.Rebind(stmt)
	// get the ids of newly inserted milestones
	milestonesInsertResult, err := tx.QueryContext(ctx, stmt, valueArgs...)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	for milestonesInsertResult.Next() {
		id := 0
		err := milestonesInsertResult.Scan(&id)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		insertedMilestoneIds = append(insertedMilestoneIds, id)
	}

	milestonesInsertResult.Close()
	if err != nil {
		return nil, err
	}

	// build a map of unique skills from all the milestones
	var allSkills []string
	for i, milestone := range input.Milestones {
		if len(milestone.Skills) == 0 {
			return nil, errors.New(fmt.Sprintf("milestone %v must have atleast one skill", i))
		}
		for _, skill := range milestone.Skills {
			allSkills = append(allSkills, *skill)
		}
	}
	skillsMap, err := findOrCreateSkills(allSkills, user.Id, tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// populate the milestoneskills table with new milestone_id's and skill_id's
	var milestoneSkillsArgs []interface{}
	var milestoneSkillsQuery []string
	for i, milestone := range input.Milestones {
		milestoneID := insertedMilestoneIds[i]
		for _, skill := range milestone.Skills {
			milestoneSkillsQuery = append(milestoneSkillsQuery, "(?, ?)")
			milestoneSkillsArgs = append(milestoneSkillsArgs, skillsMap[strings.ToLower(*skill)].Id, milestoneID)
		}
	}
	newMilestoneSkillsQuery := j.db.Rebind(fmt.Sprintf(`insert into milestoneskills(skill_id, milestone_id) values %v returning id`, strings.Join(milestoneSkillsQuery, ",")))

	_, err = tx.ExecContext(ctx, newMilestoneSkillsQuery, milestoneSkillsArgs...)
	if err != nil {
		_ = tx.Rollback()
		log.Println("error while creating non existing milestoneskills")
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return j.GetById(insertedJobId)
}

func (j *JobsRepo) UpdateJob(input *gqlmodel.UpdateJobInput) (*dbmodel.Job, error) {
	panic("Not implemented")
}

func (j *JobsRepo) DeleteJob(jobId string) (*dbmodel.Job, error) {
	panic("Not implemented")
}

// Get the complete job details based on the job id
func (j *JobsRepo) GetById(jobId string) (*dbmodel.Job, error) {
	var job dbmodel.Job
	err := j.db.QueryRowx(selectJobByIdQuery, jobId).StructScan(&job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// GetByUserId returns all jobs created by that user
func (j *JobsRepo) GetByUserId(userId string) ([]*dbmodel.Job, error) {

	rows, err := j.db.Queryx(selectJobsByUserIdQuery, userId)
	if err != nil {
		return nil, err
	}

	var jobs []*dbmodel.Job
	for rows.Next() {
		var job dbmodel.Job
		rows.StructScan(&job)
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

//TODO: Refactor this.
func (j *JobsRepo) GetStatsByUserId(userId string) (*gqlmodel.UserStats, error) {
	panic("not implemented")
}

//TODO: Add sorting order functionality
func (j *JobsRepo) GetAll(filters *gqlmodel.JobsFilterInput) ([]*dbmodel.Job, error) {
	var jobSkills []string
	for _, skill := range filters.Skills {
		jobSkills = append(jobSkills, strings.ToLower(*skill))
	}

	var jobStatuses []string
	for _, status := range filters.Status {
		jobStatuses = append(jobStatuses, strings.ToLower(status.String()))
	}

	query, args, err := sqlx.In(selectAllJobsWithFiltersQuery, jobSkills, jobStatuses)
	if err != nil {
		return nil, err
	}
	query = j.db.Rebind(query)

	rows, err := j.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}

	var result []*dbmodel.Job
	for rows != nil && rows.Next() {
		var tempJob dbmodel.Job
		rows.StructScan(&tempJob)
		result = append(result, &tempJob)
	}
	return result, nil
}

func (j *JobsRepo) GetMilestonesByJobId(jobId string) ([]*dbmodel.Milestone, error) {
	rows, err := j.db.Queryx(selectMilestonesByJobId, jobId)
	if err != nil {
		return nil, err
	}

	var milestones []*dbmodel.Milestone
	for rows.Next() {
		var milestone dbmodel.Milestone
		rows.StructScan(&milestone)
		milestones = append(milestones, &milestone)
	}
	return milestones, nil
}

func (j *JobsRepo) GetMilesoneById(milestoneId string) (*dbmodel.Milestone, error) {
	var milestone dbmodel.Milestone
	err := j.db.QueryRowx(selectMilestoneByIdQuery, milestoneId).StructScan(&milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

func (j *JobsRepo) MarkJobCompleted(ctx context.Context, jobId string) (*dbmodel.Job, error) {
	// create a new transaction
	tx, err := j.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// get all the milestones for a job
	milestones, err := j.GetMilestonesByJobId(jobId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// mark all the milestones as completed
	milestoneIds := make([]string, len(milestones))
	for i, milestone := range milestones {
		milestoneIds[i] = milestone.Id
	}
	err = j.markMilestonesCompleted(milestoneIds, tx, ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// mark the job status as completed
	err = j.markJobCompleted(jobId, tx, ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	// commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Println("error while commit job status = completed")
		log.Println(err)
		return nil, err
	}

	// TODO: Make a service layer ya lazy bum

	// we get the job after committing a transaction because if you do it before it loads a job from the database
	// except the job's data is not updated yet so you still get the stale data
	updatedJob, err := j.GetById(jobId)
	if err != nil {
		return nil, err
	}
	fmt.Println(updatedJob)
	return updatedJob, nil
}

func (j *JobsRepo) markMilestonesCompleted(milestoneIds []string, tx *sqlx.Tx, ctx context.Context) error {
	stmt, args, err := sqlx.In(updateMilestoneStatusCompleted, milestoneIds)
	if err != nil {
		return err
	}

	stmt = tx.Rebind(stmt)
	_, err = tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil
	}
	return nil
}

func (j *JobsRepo) markJobCompleted(jobId string, tx *sqlx.Tx, ctx context.Context) error {
	_, err := tx.ExecContext(ctx, updateJobStatusCompleted, jobId)
	if err != nil {
		return err
	}
	return nil
}

const (
	createJobQuery = `insert into jobs(title, description, difficulty,created_by) values ($1, $2, $3, $4) returning jobs.id`

	selectJobByIdQuery            = `SELECT * FROM jobs WHERE id = $1 and is_deleted=false`
	selectJobsByUserIdQuery       = `SELECT * FROM jobs WHERE created_by = $1 and is_deleted=false`
	selectAllJobsWithFiltersQuery = `select * from jobs where jobs.id in (select milestones.job_id from globalskills 
		join milestoneskills on globalskills.id = milestoneskills.skill_id
		join milestones on milestones.id = milestoneskills.milestone_id
		where globalskills.value in (?)
		)
		and jobs.status in (?) and jobs.is_deleted = false`
	selectMilestonesByJobId  = `SELECT * FROM milestones WHERE job_id = $1 and is_deleted = false`
	selectMilestoneByIdQuery = `SELECT * FROM milestones WHERE id = $1 and is_deleted=false`

	updateMilestoneStatusCompleted = `update milestones set status = 'completed' where id in (?) and is_deleted = false`
	updateJobStatusCompleted       = `update jobs set status = 'completed' where id = $1 and is_deleted = false`
)

func getInsertMilestonesStatement(input *gqlmodel.CreateJobInput, insertedJobId, userId string) (string, []interface{}) {
	var valueStrings []string
	var valueArgs []interface{}
	for _, milestone := range input.Milestones {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, milestone.Title, milestone.Desc, milestone.Duration, milestone.Status.String(), milestone.Resolution, insertedJobId)
	}
	stmt := fmt.Sprintf("INSERT INTO milestones (title, description, duration, status, resolution, job_id) VALUES %s returning id",
		strings.Join(valueStrings, ", "))
	return stmt, valueArgs
}
