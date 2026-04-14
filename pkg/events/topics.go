package events

const (
	TopicUserRegistered        = "user.registered"
	TopicUserLoggedIn          = "user.logged_in"
	TopicContentCreated        = "content.created"
	TopicContentUpdated        = "content.updated"
	TopicContentDeleted        = "content.deleted"
	TopicContentStatusChanged  = "content.status_changed"
	TopicContentRevisionCreated = "content.revision_created"
	TopicSearchIndexed         = "search.indexed"
	TopicSummaryGenerated      = "summary.generated"
	TopicReviewApproved        = "review.approved"
	TopicReviewRejected        = "review.rejected"
	TopicAgentTaskCreated      = "agent.task.created"
	TopicAgentOutputGenerated  = "agent.output.generated"
)
