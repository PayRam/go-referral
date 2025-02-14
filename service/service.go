package service

import (
	"github.com/PayRam/go-referral/models"
	"github.com/PayRam/go-referral/request"
	"github.com/PayRam/go-referral/response"
	"github.com/shopspring/decimal"
)

// EventService handles operations related to events
type EventService interface {
	CreateEvent(project string, request request.CreateEventRequest) (*models.Event, error)
	GetEvents(req request.GetEventsRequest) ([]models.Event, int64, error)
	UpdateEvent(project, key string, req request.UpdateEventRequest) (*models.Event, error)
}

// CampaignService handles operations related to campaigns
type CampaignService interface {
	CreateCampaign(project string, req request.CreateCampaignRequest) (*models.Campaign, error)
	GetCampaigns(req request.GetCampaignsRequest) ([]models.Campaign, int64, error)
	GetTotalCampaigns(req request.GetCampaignsRequest) (int64, error)
	UpdateCampaign(project string, id uint, req request.UpdateCampaignRequest) (*models.Campaign, error)
	SetDefaultCampaign(project string, campaignID uint) (*models.Campaign, error)
	RemoveDefaultCampaign(project string, campaignID uint) (*models.Campaign, error)
	UpdateCampaignStatus(project string, campaignID uint, newStatus string) (*models.Campaign, error)
}

// MemberService handles operations related to referral codes
type MemberService interface {
	CreateMember(project string, req request.CreateMemberRequest) (*models.Member, error)
	GetMembers(req request.GetMemberRequest) ([]models.Member, int64, error)
	GetTotalMembers(req request.GetMemberRequest) (int64, error)
	UpdateMember(project, referenceID string, request request.UpdateMemberRequest) (*models.Member, error)
	UpdateMemberStatus(project, referenceID string, newStatus string) (*models.Member, error)
}

type EventLogService interface {
	CreateEventLog(project string, req request.CreateEventLogRequest) (*models.EventLog, error)
	GetEventLogs(req request.GetEventLogRequest) ([]models.EventLog, int64, error)
}

type RewardService interface {
	GetTotalRewards(request request.GetRewardRequest) (decimal.Decimal, error)
	GetRewards(req request.GetRewardRequest) ([]models.Reward, int64, error)
	GetNewReferrerCount(req request.GetRewardRequest) (int64, error)
	GetNewRefereeCount(req request.GetRewardRequest) (int64, error)
}

type AggregatorService interface {
	GetReferrerMembersStats(req request.GetMemberRequest) ([]response.ReferrerStats, int64, error)
	GetRewardsStats(req request.GetRewardRequest) ([]response.RewardStats, error)
}

type Worker interface {
	ProcessPendingEvents() error
}
