package model

import (
	dbmodel "github.com/cassini-Inner/inner-src-mgmt-go/repository/model"
	"strings"
)

type Milestone struct {
	ID          string     `json:"id"`
	JobID       string     `json:"job"`
	Title       string     `json:"title"`
	TimeCreated string     `json:"timeCreated"`
	TimeUpdated string     `json:"timeUpdated"`
	Desc        string     `json:"desc"`
	Resolution  string     `json:"resolution"`
	Duration    string     `json:"duration"`
	Status      *JobStatus `json:"status"`
	AssignedTo  string     `json:"assignedTo"`
	Skills      []*Skill   `json:"skills"`
}

func (m *Milestone) MapDbToGql(dbMilestone dbmodel.Milestone) {
	m.ID = dbMilestone.Id
	m.JobID = dbMilestone.JobId
	m.Title = dbMilestone.Title
	m.TimeCreated = dbMilestone.TimeCreated
	m.TimeUpdated = dbMilestone.TimeUpdated
	m.Desc = dbMilestone.Description
	m.Resolution = dbMilestone.Resolution
	m.Duration = dbMilestone.Duration

	status := JobStatus(strings.ToUpper(dbMilestone.Status))
	m.Status = &status
	if dbMilestone.AssignedTo.Valid {
		m.AssignedTo = dbMilestone.AssignedTo.String
	}
}
