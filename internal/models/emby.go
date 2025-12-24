package models

type EmbyUser struct {
	ID                        string `json:"Id"`
	Name                      string `json:"Name"`
	LastActivityDate          string `json:"LastActivityDate"`
	LastPlaybackCheckIn       string `json:"LastPlaybackCheckIn"`
	HasConfiguredPassword     bool   `json:"HasConfiguredPassword"`
	HasConfiguredEasyPassword bool   `json:"HasConfiguredEasyPassword"`
	Type                      string `json:"Type"`
	Policy                    struct {
		IsAdministrator                  bool          `json:"IsAdministrator"`
		IsHidden                         bool          `json:"IsHidden"`
		IsDisabled                       bool          `json:"IsDisabled"`
		MaxParentalRating                int           `json:"MaxParentalRating"`
		BlockedTags                      []string      `json:"BlockedTags"`
		EnableUserPreferenceAccess       bool          `json:"EnableUserPreferenceAccess"`
		AccessSchedules                  []interface{} `json:"AccessSchedules"`
		BlockedChannels                  []interface{} `json:"BlockedChannels"`
		BlockedTagsForRatings            []interface{} `json:"BlockedTagsForRatings"`
		BlockedVideoTypes                []interface{} `json:"BlockedVideoTypes"`
		EnablePlaybackRemuxing           bool          `json:"EnablePlaybackRemuxing"`
		EnableLiveTvManagement           bool          `json:"EnableLiveTvManagement"`
		EnableLiveTvAccess               bool          `json:"EnableLiveTvAccess"`
		EnableMediaPlayback              bool          `json:"EnableMediaPlayback"`
		EnableAudioPlaybackTranscoding   bool          `json:"EnableAudioPlaybackTranscoding"`
		EnableVideoPlaybackTranscoding   bool          `json:"EnableVideoPlaybackTranscoding"`
		EnableContentDeletion            bool          `json:"EnableContentDeletion"`
		EnableContentDeletionFromFolders []string      `json:"EnableContentDeletionFromFolders"`
		EnableContentDownloading         bool          `json:"EnableContentDownloading"`
		EnableSyncTranscoding            bool          `json:"EnableSyncTranscoding"`
		EnableMediaConversion            bool          `json:"EnableMediaConversion"`
		EnabledDevices                   []interface{} `json:"EnabledDevices"`
		EnableAllDevices                 bool          `json:"EnableAllDevices"`
		EnabledChannels                  []interface{} `json:"EnabledChannels"`
		EnableAllChannels                bool          `json:"EnableAllChannels"`
		EnabledFolders                   []string      `json:"EnabledFolders"`
		EnableAllFolders                 bool          `json:"EnableAllFolders"`
		InvalidLoginAttemptCount         int           `json:"InvalidLoginAttemptCount"`
		LoginAttemptsBeforeLockout       int           `json:"LoginAttemptsBeforeLockout"`
		MaxActiveSessions                int           `json:"MaxActiveSessions"`
		EnablePublicSharing              bool          `json:"EnablePublicSharing"`
		RemoteClientBitrateLimit         int           `json:"RemoteClientBitrateLimit"`
		AuthenticationProviderId         string        `json:"AuthenticationProviderId"`
		PasswordResetProviderId          string        `json:"PasswordResetProviderId"`
		SyncPlayAccess                   string        `json:"SyncPlayAccess"`
	} `json:"Policy"`
}
