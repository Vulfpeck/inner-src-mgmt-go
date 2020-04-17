package postgres

import (
	gqlmodel "github.com/cassini-Inner/inner-src-mgmt-go/graph/model"
	dbmodel "github.com/cassini-Inner/inner-src-mgmt-go/postgres/models"
	"github.com/jmoiron/sqlx"
)

type JobsRepo struct {
	db *sqlx.DB
}

func NewJobsRepo(db *sqlx.DB) *JobsRepo {
	return &JobsRepo{db: db}
}

func (j *JobsRepo) CreateJob(input *gqlmodel.CreateJobInput) (*dbmodel.Job, error) {
	// query := "INSERT INTO jobs (CreatedBy, Title, Description, Difficulty, Status, TimeCreated, TimeUpdated, IsDeleted) VALUES ("
	// err := j.db.QueryRowx()
	panic("Not implemented")
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
	query := "SELECT * FROM jobs WHERE id = $1"
	err := j.db.QueryRowx(query, jobId).StructScan(&job)
	return &job, err
}

// GetByUserId returns all jobs created by that user
func (j *JobsRepo) GetByUserId(userId string) ([]*dbmodel.Job, error) {
	var job *dbmodel.Job
	var jobs []*dbmodel.Job
	query := "SELECT * FROM jobs WHERE created_by = $1" 
	rows, err := j.db.Queryx(query, userId)
	for rows.Next() {
		rows.StructScan(&job)
		jobs = append(jobs, job)
	}
	return jobs, err
}

//TODO: Refactor this.
func (j *JobsRepo) GetStatsByUserId(userId string) (*gqlmodel.UserStats, error) {
	panic("not implemented")
}

func (j *JobsRepo) GetAll(filters *gqlmodel.JobsFilterInput) ([]*dbmodel.Job, error) {
	panic("not implemented")
}
