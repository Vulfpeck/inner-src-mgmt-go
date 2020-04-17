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
func (d *DiscussionsRepo) CreateComment(jobId string, comment string) (*gqlmodel.Comment, error) {
	panic("not implemented")
}
func (d *DiscussionsRepo) UpdateComment(commentId string, comment string) (*gqlmodel.Comment, error) {
	panic("not implemented")
}
func (d *DiscussionsRepo) DeleteComment(commentId string) (*gqlmodel.Comment, error) {
	panic("not implemented")
}

func (d *DiscussionsRepo) GetByJobId(jobId string) ([]*dbmodel.Discussion, error) {
	var discussion dbmodel.Discussion
	var result []*dbmodel.Discussion

	rows, err := d.db.Queryx(getDiscussionByJobId, jobId)
	if err != nil {
		return nil, err
	}
	for rows != nil && rows.Next() {
		rows.StructScan(&discussion)
		result = append(result, &discussion)
	}

	return result, nil
}

const (
	getDiscussionByJobId = `select * from discussions where job_id = $1 order by time_created`
)
