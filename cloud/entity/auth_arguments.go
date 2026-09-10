package entity

// CloudAuthType defines the type of cloud authentication to be used.
// It's an integer type, making it lightweight and suitable for comparison and indexing.
type CloudAuthType int

const (
	// AuthTypeAccessKey indicates authentication via Access Key.
	// This method usually requires an Access Key ID and a Secret Access Key.
	AuthTypeAccessKey CloudAuthType = iota

	// AuthTypeIAMRole indicates authentication using an IAM Role.
	// This method relies on IAM roles for AWS services and resources.
	AuthTypeIAMRole
)

// CloudAuthArguments holds the necessary parameters for cloud authentication.
// The structure accommodates different authentication types and their respective credentials.
type CloudAuthArguments struct {
	// AuthMode specifies the type of authentication to use, represented by CloudAuthType.
	AuthMode CloudAuthType

	// AccessKey is the Access Key ID for AuthTypeAccessKey mode.
	// It's used along with SecretKey for authentication.
	AccessKey string

	// SecretKey is the Secret Access Key for AuthTypeAccessKey mode.
	// This key should be kept secure and not exposed in logs or other outputs.
	SecretKey string

	// SessionId is the session identifier for the current authentication session.
	// This is often used in temporary credentials scenario.
	SessionId string

	// IAMRole specifies the IAM role to be assumed for AuthTypeIAMRole mode.
	// This role should have the necessary permissions for the intended operations.
	IAMRole string

	// Region specifies the cloud provider's region to which the authentication is scoped.
	// This is essential for services that are region-specific.
	Region string
}
