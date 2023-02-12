package db

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
)

//    \/   \/   \/   \/   \/   \/   \/   \/   \/   \/   \/   \/   \/   \/   \/   \/   \/
//
// See the separate 'db-schema.sql' schema file, for descriptions of each of these tables
// and the fields within them.
//
//    /\   /\   /\   /\   /\   /\   /\   /\   /\   /\   /\   /\   /\   /\   /\   /\   /\

// GitopsEngineCluster is used to track clusters that host Argo CD instances
type GitopsEngineCluster struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"gitopsenginecluster"` //nolint

	PrimaryKeyID string `pg:"gitopsenginecluster_id,pk" varchar:"48"`

	SeqID int64 `pg:"seq_id"`

	// -- pointer to credentials for the cluster
	// -- Foreign key to: ClusterCredentials.clustercredentials_cred_id
	ClusterCredentialsID string `pg:"clustercredentials_id,notnull" varchar:"48"`
}

// GitopsEngineInstance is an Argo CD instance on an Argo CD cluster
type GitopsEngineInstance struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"gitopsengineinstance,alias:gei"` //nolint

	Gitopsengineinstance_id string `pg:"gitopsengineinstance_id,pk" varchar:"48"`
	SeqID                   int64  `pg:"seq_id"`

	// -- An Argo CD cluster may host multiple Argo CD instances; these fields
	// -- indicate which namespace this specific instance lives in.
	NamespaceName string `pg:"namespace_name,notnull" varchar:"48"`
	NamespaceUID  string `pg:"namespace_uid,notnull" varchar:"48"`

	// -- Reference to the Argo CD cluster containing the instance
	// -- Foreign key to: GitopsEngineCluster.gitopsenginecluster_id
	EngineClusterID string `pg:"enginecluster_id,notnull" varchar:"48"`
}

// ManagedEnvironment is an environment (eg a user's cluster, or a subset of that cluster) that they want to deploy applications to, using Argo CD
type ManagedEnvironment struct {
	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"managedenvironment,alias:me"` //nolint

	Managedenvironment_id string `pg:"managedenvironment_id,pk" varchar:"48"`
	SeqID                 int64  `pg:"seq_id"`

	// -- human readable name
	Name string `pg:"name,notnull" varchar:"256"`

	// -- pointer to credentials for the cluster
	// -- Foreign key to: ClusterCredentials.clustercredentials_cred_id
	ClusterCredentialsID string `pg:"clustercredentials_id,notnull" varchar:"48"`

	// -- CreatedOn field will tell us how old resources are
	CreatedOn time.Time `pg:"created_on"`
}

// ClusterCredentials contains the credentials required to access a K8s cluster.
// The credentials may be in one of two forms:
// 1) Kubeconfig state: Kubeconfig file, plus a reference to a specific context within the
//   - This is the same content as can be found in your local '~/.kube/config' file
//   - This is what the user would initially provide via the Service/Web UI/CLI
//   - There may be (likely is) a better way of doing this, but this works for now.
//
// 2) ServiceAccount state: A bearer token for a service account on the target cluster
//   - Same mechanism Argo CD users for accessing remote clusters
//
// You can tell which state the credentials are in, based on whether 'serviceaccount_bearer_token' is null.
//
// It is the job of the cluster agent to convert state 1 (kubeconfig) into a service account
// bearer token on the target cluster (state 2).
//   - This is the same operation as the `argocd cluster add` command, and is the same
//     technique used by Argo CD to interface with remove clusters.
//   - See https://github.com/argoproj/argo-cd/blob/a894d4b128c724129752bac9971c903ab6c650ba/cmd/argocd/commands/cluster.go#L116
type ClusterCredentials struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"clustercredentials,alias:cc"` //nolint

	// -- Primary key for the credentials (UID)
	ClustercredentialsCredID string `pg:"clustercredentials_cred_id,pk" varchar:"48"`

	SeqID int64 `pg:"seq_id"`

	// -- API URL for the cluster
	// -- Example: https://api.ci-ln-dlfw0qk-f76d1.origin-ci-int-gce.dev.openshift.com:6443
	Host string `pg:"host" varchar:"512"`

	// -- State 1) kube_config containing a token to a service account that has the permissions we need.
	KubeConfig string `pg:"kube_config" varchar:"65000"`

	// -- State 1) The name of a context within the kube_config
	KubeConfig_context string `pg:"kube_config_context" varchar:"64"`

	// -- State 2) ServiceAccount bearer token from the target manager cluster
	ServiceAccountBearerToken string `pg:"serviceaccount_bearer_token" varchar:"2048"`

	// -- State 2) The namespace of the ServiceAccount
	ServiceAccountNs string `pg:"serviceaccount_ns" varchar:"128"`

	// -- Indicates that ArgoCD/GitOps Service should not check the TLS certificate.
	AllowInsecureSkipTLSVerify bool `pg:"allowinsecure_skiptlsverify"`
}

// ClusterUser is an individual user/customer
// Note: This is basically placeholder: a real implementation would need to be way more complex.
type ClusterUser struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"clusteruser,alias:cu"` //nolint

	ClusterUserID string `pg:"clusteruser_id,pk" varchar:"48"`
	UserName      string `pg:"user_name,,notnull" varchar:"256"`
	SeqID         int64  `pg:"seq_id"`
}

type ClusterAccess struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"clusteraccess"` //nolint

	// -- Describes whose managed environment this is (UID)
	// -- Foreign key to: ClusterUser.ClusterUserID
	ClusterAccessUserID string `pg:"clusteraccess_user_id,pk" varchar:"48"`

	// -- Describes which managed environment the user has access to (UID)
	// -- Foreign key to: ManagedEnvironment.Managedenvironment_id
	ClusterAccessManagedEnvironmentID string `pg:"clusteraccess_managed_environment_id,pk" varchar:"48"`

	// -- Which Argo CD instance is managing the environment?
	// -- Foreign key to: GitOpsEngineInstance.Gitopsengineinstance_id
	ClusterAccessGitopsEngineInstanceID string `pg:"clusteraccess_gitops_engine_instance_id,pk" varchar:"48"`

	SeqID int64 `pg:"seq_id"`
}

type OperationState string

const (
	OperationState_Waiting     OperationState = "Waiting"
	OperationState_In_Progress OperationState = "In_Progress"
	OperationState_Completed   OperationState = "Completed"
	OperationState_Failed      OperationState = "Failed"
)

type OperationResourceType string

const (
	OperationResourceType_ManagedEnvironment    OperationResourceType = "ManagedEnvironment"
	OperationResourceType_SyncOperation         OperationResourceType = "SyncOperation"
	OperationResourceType_Application           OperationResourceType = "Application"
	OperationResourceType_RepositoryCredentials OperationResourceType = "RepositoryCredentials"
)

// Operation
// Operations are used by the backend to communicate database changes to the cluster-agent.
// It is the responsibility of the cluster agent to respond to operations, to read the database
// to discover what database changes occurred, and to ensure that Argo CD is consistent with
// the database state.
//
// See https://docs.google.com/document/d/1e1UwCbwK-Ew5ODWedqp_jZmhiZzYWaxEvIL-tqebMzo/edit#heading=h.9tzaobsoav27
// for description of Operation
type Operation struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"operation,alias:op"` //nolint

	// Auto-generated primary key, based on a random UID
	Operation_id string `pg:"operation_id,pk" varchar:"48"`

	// -- Specifies which Argo CD instance this operation is targeting
	// -- Foreign key to: GitopsEngineInstance.gitopsengineinstance_id
	InstanceID string `pg:"instance_id,notnull" varchar:"48"`

	// Primary key of the resource that was updated
	ResourceID string `pg:"resource_id,notnull" varchar:"48"`

	// -- The user that initiated the operation.
	OperationOwnerUserID string `pg:"operation_owner_user_id" varchar:"48"`

	// Resource type of the the resource that was updated.
	// This value lets the operation know which table contains the resource.
	//
	// Possible values:
	// * ClusterAccess (specified when we want Argo CD to C/R/U/D a user's cluster credentials)
	// * GitopsEngineInstance (specified to CRUD an Argo instance, for example to create a new namespace and put Argo CD in it, then signal when it's done)
	// * Application (user creates a new Application via service/web UI)
	// * RepositoryCredentials (user provides private repository credentials via web UI)
	// * SyncOperation (specified when user wants to sync an Argo CD Application)
	ResourceType OperationResourceType `pg:"resource_type,notnull" varchar:"32"`

	// -- When the operation was created. Used for garbage collection, as operations should be short lived.
	CreatedOn time.Time `pg:"created_on,notnull"`

	// -- last_state_update is set whenever state changes
	// -- (initial value should be equal to created_on)
	LastStateUpdate time.Time `pg:"last_state_update,notnull"`

	// Whether the Operation is in progress/has completed/has been processed/etc.
	// (possible values: Waiting / In_Progress / Completed / Failed)
	State OperationState `pg:"state,notnull" varchar:"30"`

	// -- If there is an error message from the operation, it is passed via this field.
	HumanReadableState string `pg:"human_readable_state" varchar:"1024"`

	SeqID int64 `pg:"seq_id"`

	// -- Amount of time to wait in seconds after last_state_update for a completed/failed operation to be garbage collected.
	GCExpirationTime int `pg:"gc_expiration_time"`
}

// Application represents an Argo CD Application CR within an Argo CD namespace.
type Application struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"application"` //nolint

	// primary key: auto-generated random uid.
	ApplicationID string `pg:"application_id,pk,notnull" varchar:"48"`

	// Name of the Application CR within the Argo CD namespace
	// Value: gitopsdepl-(uid of the gitopsdeployment)
	// Example: gitopsdepl-ac2efb8e-2e2a-45a2-9c08-feb0e2e0e29b
	Name string `pg:"name,notull" varchar:"256"`

	// '.spec' field of the Application CR
	// Note: Rather than converting individual JSON fields into SQL Table fields, we just pull the whole spec field.
	SpecField string `pg:"spec_field,notnull" varchar:"16384"`

	// Which Argo CD instance it's hosted on
	EngineInstanceInstID string `pg:"engine_instance_inst_id,notnull" varchar:"48"`

	// Which managed environment it is targetting
	// Foreign key to ManagedEnvironment.Managedenvironment_id
	Managed_environment_id string `pg:"managed_environment_id" varchar:"48"`

	SeqID int64 `pg:"seq_id"`

	// -- CreatedOn field will tell us how old resources are
	CreatedOn time.Time `pg:"created_on,notnull"`
}

// ApplicationState is the Argo CD health/sync state of the Application
type ApplicationState struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"applicationstate"` //nolint

	// -- Foreign key to Application.application_id
	Applicationstate_application_id string `pg:"applicationstate_application_id,pk" varchar:"48"`

	// -- Possible values:
	// -- * Healthy
	// -- * Progressing
	// -- * Degraded
	// -- * Suspended
	// -- * Missing
	// -- * Unknown
	Health string `pg:"health,notnull" varchar:"30"`

	// -- Possible values:
	// -- * Synced
	// -- * OutOfSync
	// -- * Unknown
	SyncStatus string `pg:"sync_status,notnull" varchar:"30"`

	Message string `pg:"message" varchar:"1024"`

	Revision string `pg:"revision" varchar:"1024"`

	Resources []byte `pg:"resources"`

	// -- human_readable_health ( 512 ) NOT NULL,
	// -- human_readable_sync ( 512 ) NOT NULL,
	// -- human_readable_state ( 512 ) NOT NULL,

	ReconciledState string `pg:"reconciled_state" varchar:"4096"`
	SyncError       string `pg:"sync_error" varchar:"4096"`
}

// DeploymentToApplicationMapping represents relationship from GitOpsDeployment CR in the namespace, to an Application table row
// This means: if we see a change in a GitOpsDeployment CR, we can easily find the corresponding database entry
// Also: if we see a change to an Argo CD Application, we can easily find the corresponding GitOpsDeployment CR
//
// See for details:
// 'What are the DeploymentToApplicationMapping, KubernetesToDBResourceMapping, and APICRToDatabaseMapping, database tables for?:
// (https://docs.google.com/document/d/1e1UwCbwK-Ew5ODWedqp_jZmhiZzYWaxEvIL-tqebMzo/edit#heading=h.45brv1rx6wmo)
type DeploymentToApplicationMapping struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"deploymenttoapplicationmapping,alias:dta"` //nolint

	// UID of GitOpsDeployment resource in K8s/KCP namespace
	// (value from '.metadata.uid' field of GitOpsDeployment)
	Deploymenttoapplicationmapping_uid_id string `pg:"deploymenttoapplicationmapping_uid_id,pk,notnull" varchar:"48"`

	// Name of the GitOpsDeployment in the namespace
	DeploymentName string `pg:"name" varchar:"256"`

	// Namespace of the GitOpsDeployment
	DeploymentNamespace string `pg:"namespace" varchar:"96"`

	// UID (.metadata.uid) of the Namespace, containing the GitOpsDeployments
	// value: (uid of namespace)
	NamespaceUID string `pg:"namespace_uid" varchar:"48"`

	// Reference to the corresponding Application row
	// -- Foreign key to: Application.ApplicationID
	ApplicationID string `pg:"application_id,notnull" varchar:"48"`

	SeqID int64 `pg:"seq_id"`
}

// APICRToDatabaseMapping_ResourceType: see 'db-schema.sql' for a description of these values.
type APICRToDatabaseMapping_ResourceType string

const (
	APICRToDatabaseMapping_ResourceType_GitOpsDeploymentManagedEnvironment   APICRToDatabaseMapping_ResourceType = "GitOpsDeploymentManagedEnvironment"
	APICRToDatabaseMapping_ResourceType_GitOpsDeploymentSyncRun              APICRToDatabaseMapping_ResourceType = "GitOpsDeploymentSyncRun"
	APICRToDatabaseMapping_ResourceType_GitOpsDeploymentRepositoryCredential APICRToDatabaseMapping_ResourceType = "GitOpsDeploymentRepositoryCredential"
)

// APICRToDatabaseMapping_DBRelationType: see 'db-schema.sql' for a description of these values.
type APICRToDatabaseMapping_DBRelationType string

const (
	APICRToDatabaseMapping_DBRelationType_ManagedEnvironment   APICRToDatabaseMapping_DBRelationType = "ManagedEnvironment"
	APICRToDatabaseMapping_DBRelationType_SyncOperation        APICRToDatabaseMapping_DBRelationType = "SyncOperation"
	APICRToDatabaseMapping_DBRelationType_RepositoryCredential APICRToDatabaseMapping_DBRelationType = "RepositoryCredential"
)

// APICRToDatabaseMapping maps API custom resources on the workspace (such as GitOpsDeploymentSyncRun), to a corresponding entry in the database.
// This allows us to quickly go from API CR <-to-> Database entry, and also to identify database entries even when the API CR has been
// deleted from the API namespace.
//
// See for details:
// 'What are the DeploymentToApplicationMapping, KubernetesToDBResourceMapping, and APICRToDatabaseMapping, database tables for?:
// (https://docs.google.com/document/d/1e1UwCbwK-Ew5ODWedqp_jZmhiZzYWaxEvIL-tqebMzo/edit#heading=h.45brv1rx6wmo)
type APICRToDatabaseMapping struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"apicrtodatabasemapping,alias:atdbm"` //nolint

	APIResourceType APICRToDatabaseMapping_ResourceType `pg:"api_resource_type,notnull" varchar:"64"`
	APIResourceUID  string                              `pg:"api_resource_uid,notnull" varchar:"64"`

	APIResourceName      string `pg:"api_resource_name,notnull" varchar:"256"`
	APIResourceNamespace string `pg:"api_resource_namespace,notnull" varchar:"256"`
	NamespaceUID         string `pg:"api_resource_namespace_uid,notnull" varchar:"64"`

	DBRelationType APICRToDatabaseMapping_DBRelationType `pg:"db_relation_type,notnull" varchar:"32"`
	DBRelationKey  string                                `pg:"db_relation_key,notnull" varchar:"64"`

	SeqID int64 `pg:"seq_id"`
}

// KubernetesToDBResourceMapping represents a generic relationship between Kubernetes CR <-> Database table
// The Kubernetes CR can be either in the workspace, or in/on a GitOpsEngine cluster namespace.
//
// Example: when the cluster agent sees an Argo CD Application CR change within a namespace, it needs a way
// to know which GitOpsEngineInstance database entries corresponds to the Argo CD namespace.
// For this we would use:
// - kubernetes_resource_type: Namespace
// - kubernetes_resource_uid: (uid of namespace)
// - db_relation_type: GitOpsEngineInstance
// - db_relation_key: (primary key of gitops engine instance)
//
// Later, we can query this table to go from 'argo cd instance namespace' <= to => 'GitopsEngineInstance database row'
//
// See DeploymentToApplicationMapping for another example of this.
//
// This is also useful for tracking the lifecycle between CRs <-> database table.
type KubernetesToDBResourceMapping struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"kubernetestodbresourcemapping,alias:ktdbrm"` //nolint

	KubernetesResourceType string `pg:"kubernetes_resource_type,pk,notnull" varchar:"64"`

	KubernetesResourceUID string `pg:"kubernetes_resource_uid,pk,notnull" varchar:"64"`

	DBRelationType string `pg:"db_relation_type,pk,notnull" varchar:"64"`

	DBRelationKey string `pg:"db_relation_key,pk,notnull" varchar:"64"`

	SeqID int64 `pg:"seq_id"`
}

// SyncOperation tracks a sync request from the API. This will correspond to a sync operation on an Argo CD Application, which
// will cause Argo CD to deploy the K8s resources from Git, to the target environment. This is also known as manual sync.
type SyncOperation struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"syncoperation,alias:so"` //nolint

	SyncOperationID string `pg:"syncoperation_id,pk,notnull" varchar:"48"`

	ApplicationID string `pg:"application_id" varchar:"48"`

	DeploymentNameField string `pg:"deployment_name,notnull" varchar:"256"`

	Revision string `pg:"revision,notnull" varchar:"256"`

	DesiredState string `pg:"desired_state,notnull" varchar:"16"`

	CreatedOn time.Time `pg:"created_on,notnull"`
}

// DisposableResource can be implemented by a type, such that calling Dispose(...) on an instance of that type will delete
// the corresponding row from the database.
//
// This is an optional interface, for convenience purposes.
type DisposableResource interface {
	Dispose(ctx context.Context, dbq DatabaseQueries) error
}

// AppScopedDisposableResource can be implemented by an application-scoped type, such they calling Dispose() on an instance
// of that type will delete the row from the database.
//
// This is an optional interface, for convenience purposes.
type AppScopedDisposableResource interface {
	DisposeAppScoped(ctx context.Context, dbq ApplicationScopedQueries) error
}

// RepositoryCredentials represents a RepositoryCredentials CR.
// It is created by the backend component, if we need to access a private repository.
// Can be used as a reference via the Operation row by providing the ResourceID and ResourceType.
type RepositoryCredentials struct {

	//lint:ignore U1000 used by go-pg
	tableName struct{} `pg:"repositorycredentials,alias:rc"` //nolint

	// RepositoryCredentialsID is the PK (Primary Key) from the database, that is an auto-generated random UID.
	RepositoryCredentialsID string `pg:"repositorycredentials_id,pk,notnull" varchar:"48"`

	// UserID represents a customer of the GitOps service that wants to use a private repository.
	// -- Foreign key to: ClusterUser.ClusterUserID
	UserID string `pg:"repo_cred_user_id,notnull" varchar:"48"`

	// PrivateURL is the address of the private Git repository.
	PrivateURL string `pg:"repo_cred_url,notnull" varchar:"512"`

	// AuthUsername is the authorized username login for accessing the private Git repo.
	AuthUsername string `pg:"repo_cred_user" varchar:"256"`

	// AuthPassword is the authorized password login for accessing the private Git repo (usually encoded with Base64).
	AuthPassword string `pg:"repo_cred_pass" varchar:"1024"`

	// AuthSSHKey (alternative authentication method) is the authorized private SSH key
	// that provides access to the private Git repo. It can also be used for decrypting Sealed secrets.
	AuthSSHKey string `pg:"repo_cred_ssh" varchar:"1024"`

	// SecretObj is the name of the (insecure and unencrypted) Kubernetes secret object that provides
	// the credentials (AuthUsername & AuthPassword, OR the AuthSSHKey) to the GitOps Engine (e.g. ArgoCD)
	// to gain access into the PrivateURL repo.
	SecretObj string `pg:"repo_cred_secret,notnull" varchar:"48"`

	// EngineClusterID is the internal RedHat Managed cluster where the GitOps Engine (e.g. ArgoCD) is running.
	// -- NOTE: It is expected the SecretObj to be stored there as well.
	// -- Foreign key to: GitopsEngineInstance.Gitopsengineinstance_id
	EngineClusterID string `pg:"repo_cred_engine_id,notnull" varchar:"48"`

	// SeqID is used only for debugging purposes. It helps us to keep track of the order that rows are created.
	SeqID int64 `pg:"seq_id"`

	// -- CreatedOn field will tell us how old resources are
	CreatedOn time.Time `pg:"created_on,notnull"`
}

// hasEmptyValues returns error if any of the notnull tagged fields are empty.
func (rc *RepositoryCredentials) hasEmptyValues(fieldNamesToIgnore ...string) error {
	s := reflect.ValueOf(rc).Elem()
	typeOfObj := s.Type()

outer_for:
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		tag := typeOfObj.Field(i).Tag.Get("pg")

		// Check fields tagged with `notnull` and throw error if they are null (empty)
		if strings.Contains(tag, "notnull") {
			if f.Interface() == reflect.Zero(f.Type()).Interface() {
				fieldName := typeOfObj.Field(i).Name

				// If the field is on the list of fields to ignore, then skip to the next field on match.
				for _, fieldNameToIgnore := range fieldNamesToIgnore {
					if fieldName == fieldNameToIgnore {
						continue outer_for
					}
				}

				return fmt.Errorf("%s.%s is empty, but it shouldn't (notnull tag found: `%s`)", typeOfObj.Name(), fieldName, tag)
			}
		}
	}
	return nil
}

func (o Operation) GetGCExpirationTime() time.Duration {
	return time.Duration(o.GCExpirationTime) * time.Second
}
