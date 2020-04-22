package resolver

import (
	"context"
	gqlmodel "github.com/cassini-Inner/inner-src-mgmt-go/graph/model"
)

func (r *jobResolver) CreatedBy(ctx context.Context, obj *gqlmodel.Job) (*gqlmodel.User, error) {
	return getUserLoader(ctx).Load(obj.CreatedBy)
}

func (r *jobResolver) Discussion(ctx context.Context, obj *gqlmodel.Job) (*gqlmodel.Discussions, error) {
	discussionsList, err := r.DiscussionsRepo.GetByJobId(obj.ID)

	var commentsList []*gqlmodel.Comment
	if err != nil {
		return nil, err
	}
	for _, discussion := range discussionsList {
		var comment gqlmodel.Comment
		comment.MapDbToGql(*discussion)
		commentsList = append(commentsList, &comment)
	}
	commentsLength := len(commentsList)
	return &gqlmodel.Discussions{Discussions: commentsList, TotalCount: &commentsLength}, nil
}

//Get the list of milestones in dbmodel type, converts it to gqlmodel type and returns list of milestones
func (r *jobResolver) Milestones(ctx context.Context, obj *gqlmodel.Job) (*gqlmodel.Milestones, error) {
	milestones, err := getMilestoneByJobIdLoader(ctx).Load(obj.ID)
	if err != nil {
		return nil, err
	}
	totalCount := len(milestones)
	return &gqlmodel.Milestones{
		TotalCount: &totalCount,
		Milestones: milestones,
	}, nil
}

func (r *jobResolver) Skills(ctx context.Context, obj *gqlmodel.Job) ([]*gqlmodel.Skill, error) {
	skills, err := r.SkillsRepo.GetByJobId(obj.ID)
	if err != nil {
		return nil, err
	}
	var result []*gqlmodel.Skill
	for _, skill := range skills {
		var gqlskill gqlmodel.Skill
		gqlskill.MapDbToGql(*skill)
		result = append(result, &gqlskill)
	}
	return result, nil
}

func (r *jobResolver) Applications(ctx context.Context, obj *gqlmodel.Job) (*gqlmodel.Applications, error) {
	applications, err := r.ApplicationsRepo.GetByJobId(obj.ID)
	if err != nil {
		return nil, err
	}

	var gqlApplicationsList []*gqlmodel.Application
	for _, application := range applications {
		var gqlApplication gqlmodel.Application
		gqlApplication.MapDbToGql(*application)
		gqlApplicationsList = append(gqlApplicationsList, &gqlApplication)
	}

	//TODO: Implement the counters
	return &gqlmodel.Applications{Applications: gqlApplicationsList}, nil
}
