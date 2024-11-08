package services

type VeeamConfig struct {
	OrganizationName       string
	BackupServerUid        string
	RepositoryUid          string
	QuotaGb                float32
	RepositoryFriendlyName string
	TemplateJobUid         string
	JobSchedulerType       string
	HighPriorityJob        bool
	HostUid                string
}
