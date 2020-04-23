package postgres

import (
	gqlmodel "github.com/cassini-Inner/inner-src-mgmt-go/graph/model"
	dbmodel "github.com/cassini-Inner/inner-src-mgmt-go/postgres/model"
	"github.com/jmoiron/sqlx"
)

type DiscussionsRepo struct {
	db *sqlx.DB
}

func NewDiscussionsRepo(db *sqlx.DB) *DiscussionsRepo {
	return &DiscussionsRepo{db: db}
}

//TODO: Implement
func (d *DiscussionsRepo) CreateComment(jobId, comment, userId string) (*dbmodel.Discussion, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	insertedCommentId := 0
	err = tx.QueryRow(`insert into discussions(job_id, created_by, content) values ($1,$2, $3) returning id`, jobId, userId, comment).Scan(&insertedCommentId)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	var id,job, createdBy, content, timeCreated, timeUpdated string
	err = tx.QueryRow(`select id, job_id, created_by,content, time_created, time_updated from discussions where id = $1`, insertedCommentId).Scan(&id, &job, &createdBy, &content, &timeCreated, &timeUpdated)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &dbmodel.Discussion{
		Id:          id,
		JobId:       job,
		CreatedBy:   createdBy,
		Content:     content,
		TimeCreated: timeCreated,
		TimeUpdated: timeUpdated,
	}, nil
}
func (d *DiscussionsRepo) UpdateComment(commentId string, comment string) (*gqlmodel.Comment, error) {
	panic("not implemented")
}
func (d *DiscussionsRepo) DeleteComment(commentId string) (*gqlmodel.Comment, error) {
	panic("not implemented")
}

func (d *DiscussionsRepo) GetByJobId(jobId string) ([]*dbmodel.Discussion, error) {
	rows, err := d.db.Queryx(getDiscussionByJobId, jobId)
	if err != nil {
		return nil, err
	}

	var result []*dbmodel.Discussion
	for rows != nil && rows.Next() {
		var discussion dbmodel.Discussion
		rows.StructScan(&discussion)
		result = append(result, &discussion)
	}

	return result, nil
}

const (
	getDiscussionByJobId = `select * from discussions where job_id = $1 and is_deleted=false order by time_created`
)
