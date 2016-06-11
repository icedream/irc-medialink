package soundcloud

import ()

type v2Group struct {
	v2Object
	ArtworkURL       string     `json:"artwork_url"`
	CreatedAt        timeString `json:"created_at"`
	Creator          v2User     `json:"creator"`
	Description      string     `json:"description"`
	MembersCount     uint64     `json:"members_count"`
	Moderated        bool       `json:"moderated"`
	Name             string     `json:"name"`
	Permalink        string     `json:"permalink"`
	PermalinkURL     string     `json:"permalink_url"`
	ShortDescription string     `json:"short_description"`
	TrackCount       uint64     `json:"track_count"`
	URI              string     `json:"uri"`
}

type v2Object struct {
	ID   uint64 `json:"id"`
	Kind v2Kind `json:"kind"`
}

type v2Playlist struct {
	v2Object
	ArtworkURL     string     `json:"artwork_url"`
	CreatedAt      timeString `json:"created_at"`
	Description    string     `json:"description"`
	Duration       uint64     `json:"duration"`
	EmbeddableBy   string     `json:"embeddable_by"`
	Genre          string     `json:"genre"`
	IsAlbum        bool       `json:"is_album"`
	LabelName      string     `json:"label_name"`
	LastModified   timeString `json:"last_modified"`
	License        string     `json:"license"`
	LikesCount     uint64     `json:"likes_count"`
	ManagedByFeeds bool       `json:"managed_by_feeds"`
	Permalink      string     `json:"permalink"`
	PermalinkURL   string     `json:"permalink_url"`
	Public         bool       `json:"public"`
	PublishedAt    timeString `json:"published_at"`
	PurchaseTitle  string     `json:"purchase_title"`
	PurchaseURL    string     `json:"purchase_url"`
	ReleaseDate    timeString `json:"release_date"`
	RepostsCount   uint64     `json:"reposts_count"`
	SecretToken    string     `json:"secret_token"`
	SetType        string     `json:"set_type"`
	Sharing        string     `json:"sharing"`
	TagList        string     `json:"tag_list"`
	Title          string     `json:"title"`
	TrackCount     uint64     `json:"track_count"`
	URI            string     `json:"uri"`
	User           v2User     `json:"user"`
	UserID         uint64     `json:"user_id"`
}

type v2Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type v2Subscription struct {
	Product   v2Product `json:"product"`
	Recurring bool      `json:"recurring"`
}

type v2Track struct {
	v2Object
	ArtworkURL        string     `json:"artwork_url"`
	CommentCount      uint64     `json:"comment_count"`
	Commentable       bool       `json:"commentable"`
	CreatedAt         timeString `json:"created_at"`
	Description       string     `json:"description"`
	DownloadCount     uint64     `json:"download_count"`
	DownloadURL       string     `json:"download_url"`
	Downloadable      bool       `json:"downloadable"`
	Duration          uint64     `json:"duration"`
	EmbeddableBy      string     `json:"embeddable_by"`
	FullDuration      uint64     `json:"full_duration"`
	Genre             string     `json:"genre"`
	HasDownloadsLeft  bool       `json:"has_downloads_left"`
	LabelName         string     `json:"label_name"`
	LastModified      timeString `json:"last_modified"`
	License           string     `json:"license"`
	LikesCount        uint64     `json:"likes_count"`
	MonetizationModel string     `json:"monetization_model"`
	Permalink         string     `json:"permalink"`
	PermalinkURL      string     `json:"permalink_url"`
	PlaybackCount     uint64     `json:"playback_count"`
	Policy            string     `json:"policy"`
	Public            bool       `json:"public"`
	PurchaseTitle     string     `json:"purchase_title"`
	PurchaseURL       string     `json:"purchase_url"`
	ReleaseDate       timeString `json:"release_date"`
	RepostsCount      uint64     `json:"reposts_count"`
	SecretToken       string     `json:"secret_token"`
	Sharing           string     `json:"sharing"`
	State             string     `json:"state"`
	Streamable        bool       `json:"streamable"`
	TagList           string     `json:"tag_list"`
	Title             string     `json:"title"`
	URI               string     `json:"uri"`
	URN               string     `json:"urn"`
	User              v2User     `json:"user"`
	UserID            uint64     `json:"user_id"`
	Visuals           v2Visuals  `json:"visuals"`
	WaveformURL       string     `json:"waveform_url"`
}

type v2User struct {
	v2Object
	AvatarURL       string     `json:"avatar_url"`
	City            string     `json:"city"`
	CommentsCount   uint64     `json:"comments_count"`
	CountryCode     string     `json:"country_code"`
	Description     string     `json:"description"`
	FirstName       string     `json:"first_name"`
	FollowersCount  uint64     `json:"followers_count"`
	FollowingsCount uint64     `json:"followings_count"`
	FullName        string     `json:"full_name"`
	GroupsCount     uint64     `json:"groups_count"`
	LastModified    timeString `json:"last_modified"`
	LastName        string     `json:"last_name"`
	LikesCount      uint64     `json:"likes_count"`
	Permalink       string     `json:"permalink"`
	PermalinkURL    string     `json:"permalink_url"`
	PlaylistCount   uint64     `json:"playlist_count"`
	RepostsCount    uint64     `json:"reposts_count"`
	TrackCount      uint64     `json:"track_count"`
	URI             string     `json:"uri"`
	URN             string     `json:"urn"`
	Username        string     `json:"username"`
	Verified        bool       `json:"verified"`
	Visuals         v2Visuals  `json:"visuals"`
}

type v2VisualItem struct {
	EntryTime uint64 `json:"entry_time"`
	URN       string `json:"urn"`
	VisualURL string `json:"visual_url"`
}

type v2Visuals struct {
	Enabled bool   `json:"enabled"`
	URN     string `json:"urn"`
}
