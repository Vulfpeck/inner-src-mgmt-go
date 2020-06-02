package impl

import (
	"context"
	dbmodel "github.com/cassini-Inner/inner-src-mgmt-go/repository/model"
	"github.com/jmoiron/sqlx"
)

type DiscussionsRepoImpl struct {
	db *sqlx.DB
}

func NewDiscussionsRepo(db *sqlx.DB) *DiscussionsRepoImpl {
	return &DiscussionsRepoImpl{db: db}
}

func (d *DiscussionsRepoImpl) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return d.db.BeginTxx(ctx, nil)
}

func (d *DiscussionsRepoImpl) CommitTx(ctx context.Context, tx *sqlx.Tx) (err error) {
	err = tx.Commit()
	if err != nil {
		err = tx.Rollback()
	}
	return err
}

func (d *DiscussionsRepoImpl) CreateComment(ctx context.Context, tx *sqlx.Tx, jobId, comment, userId string) (*dbmodel.Discussion, error) {
	var newDiscussion dbmodel.Discussion
	err := tx.QueryRowxContext(ctx, `insert into discussions(job_id, created_by, content) values ($1,$2, $3) returning *`, jobId, userId, comment).StructScan(&newDiscussion)
	if err != nil {
		return nil, err
	}
	return &newDiscussion, nil
}

func (d *DiscussionsRepoImpl) UpdateComment(ctx context.Context, tx *sqlx.Tx, discussionId, content string) (*dbmodel.Discussion, error) {
	var discussion dbmodel.Discussion

	err := tx.QueryRowxContext(ctx, updateDiscussionById, content, discussionId).StructScan(&discussion)

	if err != nil {
		return nil, err
	}

	return &discussion, nil
}

func (d *DiscussionsRepoImpl) DeleteComment(ctx context.Context, tx *sqlx.Tx, discussionId string) (*dbmodel.Discussion, error) {
	var discussion dbmodel.Discussion
	err := tx.QueryRowxContext(ctx, deleteDiscussionById, discussionId).StructScan(&discussion)
	if err != nil {
		return nil, err
	}
	return &discussion, nil
}

func (d *DiscussionsRepoImpl) GetByJobId(jobId string) ([]*dbmodel.Discussion, error) {
	rows, err := d.db.Queryx(getDiscussionsByJobId, jobId)
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

func (d *DiscussionsRepoImpl) GetById(tx *sqlx.Tx, discussionId string) (*dbmodel.Discussion, error) {
	var discussion dbmodel.Discussion
	err := tx.QueryRowx(getDiscussionById, discussionId).StructScan(&discussion)
	if err != nil {
		return nil, err
	}

	return &discussion, nil
}

func (d *DiscussionsRepoImpl) DeleteAllCommentsForJob(tx *sqlx.Tx, jobID string) error {
	_, err := tx.Exec(deleteDiscussionsForJobIdQuery, jobID)
	if err != nil {
		return err
	}

	return nil
}

const (
	getDiscussionsByJobId          = `select * from discussions where job_id = $1 and is_deleted=false order by time_created`
	getDiscussionById              = `select * from discussions where id = $1 and is_deleted = false`
	updateDiscussionById           = `update discussions set content = $1 where id = $2 and is_deleted = false returning *`
	deleteDiscussionById           = `update discussions set is_deleted = true where id = $1 returning *`
	deleteDiscussionsForJobIdQuery = `update discussions set is_deleted = true where job_id = $1`
)
