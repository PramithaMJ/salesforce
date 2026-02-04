// Package connect provides Connect REST API (Chatter) operations.
package connect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Feed represents a Chatter feed.
type Feed struct {
	Elements       []FeedElement `json:"elements"`
	NextPageUrl    string        `json:"nextPageUrl,omitempty"`
	CurrentPageUrl string        `json:"currentPageUrl"`
}

// FeedElement represents a feed element (post).
type FeedElement struct {
	ID              string       `json:"id"`
	Type            string       `json:"type"`
	URL             string       `json:"url"`
	CreatedDate     string       `json:"createdDate"`
	ModifiedDate    string       `json:"modifiedDate,omitempty"`
	Body            MessageBody  `json:"body"`
	Actor           Actor        `json:"actor"`
	Capabilities    Capabilities `json:"capabilities,omitempty"`
	Header          TextBody     `json:"header,omitempty"`
}

// MessageBody contains feed element body.
type MessageBody struct {
	MessageSegments []MessageSegment `json:"messageSegments"`
	Text            string           `json:"text"`
}

// MessageSegment represents a segment of a message.
type MessageSegment struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	Name         string `json:"name,omitempty"`
	URL          string `json:"url,omitempty"`
	RecordId     string `json:"recordId,omitempty"`
	MentionedUser *User `json:"user,omitempty"`
}

// TextBody contains text content.
type TextBody struct {
	Text string `json:"text"`
}

// Actor represents a user or entity.
type Actor struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Photo       Photo  `json:"photo,omitempty"`
	URL         string `json:"url,omitempty"`
}

// Photo contains photo URLs.
type Photo struct {
	SmallPhotoUrl  string `json:"smallPhotoUrl"`
	MediumPhotoUrl string `json:"mediumPhotoUrl,omitempty"`
	LargePhotoUrl  string `json:"largePhotoUrl,omitempty"`
	FullEmailPhotoUrl string `json:"fullEmailPhotoUrl,omitempty"`
}

// Capabilities contains feed element capabilities.
type Capabilities struct {
	Comments    CommentsCapability `json:"comments,omitempty"`
	Files       FilesCapability    `json:"files,omitempty"`
	Like        LikeCapability     `json:"like,omitempty"`
	Poll        PollCapability     `json:"poll,omitempty"`
}

// CommentsCapability contains comment capability info.
type CommentsCapability struct {
	Page CommentPage `json:"page"`
}

// CommentPage contains comments.
type CommentPage struct {
	Items          []Comment `json:"items"`
	TotalCount     int       `json:"total"`
	NextPageUrl    string    `json:"nextPageUrl,omitempty"`
}

// Comment represents a comment.
type Comment struct {
	ID          string      `json:"id"`
	Body        MessageBody `json:"body"`
	CreatedDate string      `json:"createdDate"`
	User        User        `json:"user"`
	URL         string      `json:"url"`
}

// FilesCapability contains file capability info.
type FilesCapability struct {
	Items []ContentDocument `json:"items"`
}

// LikeCapability contains like capability info.
type LikeCapability struct {
	IsLikedByCurrentUser bool `json:"isLikedByCurrentUser"`
	LikesMessage         TextBody `json:"likesMessage,omitempty"`
}

// PollCapability contains poll capability info.
type PollCapability struct {
	Choices     []PollChoice `json:"choices"`
	TotalVotes  int          `json:"totalVoteCount"`
}

// PollChoice represents a poll choice.
type PollChoice struct {
	ID         string `json:"id"`
	Position   int    `json:"position"`
	Text       string `json:"text"`
	VoteCount  int    `json:"voteCount"`
}

// User represents a Chatter user.
type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DisplayName     string `json:"displayName,omitempty"`
	Title           string `json:"title,omitempty"`
	Email           string `json:"email,omitempty"`
	CompanyName     string `json:"companyName,omitempty"`
	Username        string `json:"username,omitempty"`
	Photo           Photo  `json:"photo,omitempty"`
	URL             string `json:"url,omitempty"`
	Type            string `json:"type,omitempty"`
	IsActive        bool   `json:"isActive"`
	UserType        string `json:"userType,omitempty"`
}

// UserPage contains a page of users.
type UserPage struct {
	Users       []User `json:"users"`
	NextPageUrl string `json:"nextPageUrl,omitempty"`
	TotalCount  int    `json:"total"`
}

// ContentDocument represents a file.
type ContentDocument struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	FileType       string `json:"fileType"`
	ContentSize    int    `json:"contentSize"`
	DownloadUrl    string `json:"downloadUrl"`
	RenditionUrl   string `json:"renditionUrl,omitempty"`
	VersionId      string `json:"versionId"`
}

// FilePage contains a page of files.
type FilePage struct {
	Files       []ContentDocument `json:"files"`
	NextPageUrl string            `json:"nextPageUrl,omitempty"`
	TotalCount  int               `json:"total"`
}

// Group represents a Chatter group.
type Group struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	MemberCount     int    `json:"memberCount"`
	Owner           User   `json:"owner"`
	Visibility      string `json:"visibility"`
	CanHaveChatterGuests bool `json:"canHaveChatterGuests"`
	URL             string `json:"url"`
	Photo           Photo  `json:"photo,omitempty"`
}

// GroupPage contains a page of groups.
type GroupPage struct {
	Groups      []Group `json:"groups"`
	NextPageUrl string  `json:"nextPageUrl,omitempty"`
	TotalCount  int     `json:"total"`
}

// FeedInput represents input for creating a feed element.
type FeedInput struct {
	Body            MessageBodyInput `json:"body"`
	SubjectId       string           `json:"subjectId"`
	FeedElementType string           `json:"feedElementType,omitempty"`
	Visibility      string           `json:"visibility,omitempty"`
}

// MessageBodyInput represents input for message body.
type MessageBodyInput struct {
	MessageSegments []MessageSegmentInput `json:"messageSegments"`
}

// MessageSegmentInput represents input for a message segment.
type MessageSegmentInput struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Id   string `json:"id,omitempty"`
}

// HTTPClient interface for dependency injection.
type HTTPClient interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, body interface{}) ([]byte, error)
	Patch(ctx context.Context, path string, body interface{}) ([]byte, error)
	Delete(ctx context.Context, path string) ([]byte, error)
}

// Service provides Connect REST API operations.
type Service struct {
	client     HTTPClient
	apiVersion string
}

// NewService creates a new Connect service.
func NewService(client HTTPClient, apiVersion string) *Service {
	return &Service{client: client, apiVersion: apiVersion}
}

// GetNewsFeed retrieves the current user's news feed.
func (s *Service) GetNewsFeed(ctx context.Context) (*Feed, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/feeds/news/me/feed-elements", s.apiVersion)
	return s.getFeed(ctx, path)
}

// GetUserProfileFeed retrieves a user's profile feed.
func (s *Service) GetUserProfileFeed(ctx context.Context, userId string) (*Feed, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/feeds/user-profile/%s/feed-elements", s.apiVersion, userId)
	return s.getFeed(ctx, path)
}

// GetRecordFeed retrieves a record's feed.
func (s *Service) GetRecordFeed(ctx context.Context, recordId string) (*Feed, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/feeds/record/%s/feed-elements", s.apiVersion, recordId)
	return s.getFeed(ctx, path)
}

// GetGroupFeed retrieves a group's feed.
func (s *Service) GetGroupFeed(ctx context.Context, groupId string) (*Feed, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/feeds/groups/%s/feed-elements", s.apiVersion, groupId)
	return s.getFeed(ctx, path)
}

// GetFeedElement retrieves a single feed element.
func (s *Service) GetFeedElement(ctx context.Context, feedElementId string) (*FeedElement, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/feed-elements/%s", s.apiVersion, feedElementId)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var elem FeedElement
	if err := json.Unmarshal(respBody, &elem); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &elem, nil
}

// PostFeedElement creates a new feed element (post).
func (s *Service) PostFeedElement(ctx context.Context, input FeedInput) (*FeedElement, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/feed-elements", s.apiVersion)
	if input.FeedElementType == "" {
		input.FeedElementType = "FeedItem"
	}
	respBody, err := s.client.Post(ctx, path, input)
	if err != nil {
		return nil, err
	}
	var elem FeedElement
	if err := json.Unmarshal(respBody, &elem); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &elem, nil
}

// PostComment adds a comment to a feed element.
func (s *Service) PostComment(ctx context.Context, feedElementId string, body MessageBodyInput) (*Comment, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/feed-elements/%s/capabilities/comments/items", s.apiVersion, feedElementId)
	respBody, err := s.client.Post(ctx, path, map[string]interface{}{"body": body})
	if err != nil {
		return nil, err
	}
	var comment Comment
	if err := json.Unmarshal(respBody, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &comment, nil
}

// LikeFeedElement likes a feed element.
func (s *Service) LikeFeedElement(ctx context.Context, feedElementId string) error {
	path := fmt.Sprintf("/services/data/v%s/chatter/feed-elements/%s/capabilities/chatter-likes/items", s.apiVersion, feedElementId)
	_, err := s.client.Post(ctx, path, nil)
	return err
}

// UnlikeFeedElement unlikes a feed element.
func (s *Service) UnlikeFeedElement(ctx context.Context, feedElementId, likeId string) error {
	path := fmt.Sprintf("/services/data/v%s/chatter/likes/%s", s.apiVersion, likeId)
	_, err := s.client.Delete(ctx, path)
	return err
}

// DeleteFeedElement deletes a feed element.
func (s *Service) DeleteFeedElement(ctx context.Context, feedElementId string) error {
	path := fmt.Sprintf("/services/data/v%s/chatter/feed-elements/%s", s.apiVersion, feedElementId)
	_, err := s.client.Delete(ctx, path)
	return err
}

// GetCurrentUser retrieves the current user's info.
func (s *Service) GetCurrentUser(ctx context.Context) (*User, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/users/me", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &user, nil
}

// GetUser retrieves a user's info.
func (s *Service) GetUser(ctx context.Context, userId string) (*User, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/users/%s", s.apiVersion, userId)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &user, nil
}

// SearchUsers searches for users.
func (s *Service) SearchUsers(ctx context.Context, query string) (*UserPage, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/users?q=%s", s.apiVersion, url.QueryEscape(query))
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var page UserPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &page, nil
}

// GetMyFiles retrieves current user's files.
func (s *Service) GetMyFiles(ctx context.Context) (*FilePage, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/users/me/files", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var page FilePage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &page, nil
}

// GetFile retrieves file information.
func (s *Service) GetFile(ctx context.Context, fileId string) (*ContentDocument, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/files/%s", s.apiVersion, fileId)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var file ContentDocument
	if err := json.Unmarshal(respBody, &file); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &file, nil
}

// GetGroups retrieves Chatter groups.
func (s *Service) GetGroups(ctx context.Context) (*GroupPage, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/groups", s.apiVersion)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var page GroupPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &page, nil
}

// GetGroup retrieves a Chatter group.
func (s *Service) GetGroup(ctx context.Context, groupId string) (*Group, error) {
	path := fmt.Sprintf("/services/data/v%s/chatter/groups/%s", s.apiVersion, groupId)
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var group Group
	if err := json.Unmarshal(respBody, &group); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &group, nil
}

func (s *Service) getFeed(ctx context.Context, path string) (*Feed, error) {
	respBody, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	var feed Feed
	if err := json.Unmarshal(respBody, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &feed, nil
}
