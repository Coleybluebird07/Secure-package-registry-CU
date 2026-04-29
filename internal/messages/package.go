package messages

type PackageRequest struct {
	Ecosystem  string
	Identifier string
}

type PackageUpdated struct {
	Ecosystem  string
	Identifier string
	Version    string
}

type CollectionRequested struct {
	TaskID     int32
	Ecosystem  string
	Identifier string
	Version    string
}

type CollectionCompleted struct {
	TaskID         int32
	Ecosystem      string
	Identifier     string
	Version        string
	Success        bool
	ArtifactBucket string // empty on failure
	ArtifactKey    string // empty on failure
	FailureReason  string // empty on success
}

type RebuildRequested struct {
	TaskID     int32
	Ecosystem  string
	Identifier string
	Version    string
	Source     string
}

type RebuildCompleted struct {
	TaskID     int32
	Ecosystem  string
	Identifier string
	Version    string
	Source     string

	Success     bool
	Unavailable bool
	Matched     bool

	OfficialArtifactBucket string
	OfficialArtifactKey    string

	RebuiltArtifactBucket string
	RebuiltArtifactKey    string

	DiffoscopeBucket string
	DiffoscopeKey    string

	LogsBucket string
	LogsKey    string

	MetadataBucket string
	MetadataKey    string

	FailureReason string
}
