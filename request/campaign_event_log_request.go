package request

import "gorm.io/gorm"

type GetCampaignEventLogRequest struct {
	Projects             []string             `form:"projects"` // Filter by projects
	IDs                  []uint               `form:"ids"`      // Filter by ID
	CampaignIDs          []uint               `form:"campaignIDs"`
	EventIDs             []uint               `form:"eventIDs"`
	MemberIDs            []uint               `form:"memberIDs"`
	MemberReferenceIDs   []string             `form:"memberReferenceIDs"`
	Status               []string             `form:"status"`
	EventLogIDs          []uint               `form:"eventLogIDs"`
	ReferredRewardIDs    []uint               `form:"referredRewardIDs"`
	RefereeRewardIDs     []uint               `form:"refereeRewardIDs"`
	PaginationConditions PaginationConditions `form:"paginationConditions"` // Embedded pagination and sorting struct
}

func ApplyGetCampaignEventLogRequest(req GetCampaignEventLogRequest, query *gorm.DB) *gorm.DB {
	// Apply filters with table name prepended
	if len(req.Projects) > 0 {
		query = query.Where("referral_campaign_event_logs.project IN (?)", req.Projects)
	}
	if len(req.IDs) > 0 {
		query = query.Where("referral_campaign_event_logs.id IN (?)", req.IDs)
	}
	if len(req.CampaignIDs) > 0 {
		query = query.Where("referral_campaign_event_logs.campaign_id IN (?)", req.CampaignIDs)
	}
	if len(req.EventIDs) > 0 {
		query = query.Where("referral_campaign_event_logs.event_id IN (?)", req.EventIDs)
	}
	if len(req.MemberIDs) > 0 {
		query = query.Where("referral_campaign_event_logs.member_id IN (?)", req.MemberIDs)
	}
	if len(req.MemberReferenceIDs) > 0 {
		query = query.Where("referral_campaign_event_logs.member_reference_id IN (?)", req.MemberReferenceIDs)
	}
	if len(req.Status) > 0 {
		query = query.Where("referral_campaign_event_logs.status IN (?)", req.Status)
	}
	if len(req.EventLogIDs) > 0 {
		query = query.Where("referral_campaign_event_logs.event_log_id IN (?)", req.EventLogIDs)
	}
	if len(req.ReferredRewardIDs) > 0 {
		query = query.Where("referral_campaign_event_logs.referred_reward_id IN (?)", req.ReferredRewardIDs)
	}
	if len(req.RefereeRewardIDs) > 0 {
		query = query.Where("referral_campaign_event_logs.referee_reward_id IN (?)", req.RefereeRewardIDs)
	}

	return query
}
