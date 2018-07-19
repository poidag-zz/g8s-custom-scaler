package g8s

type Cluster struct {
	Workers []Worker `json:"workers"`
}

type Worker struct {
	Memory struct {
		SizeGb int `json:"size_gb"`
	} `json:"memory"`
	Storage struct {
		SizeGb int `json:"size_gb"`
	} `json:"storage"`
	CPU struct {
		Cores int `json:"cores"`
	} `json:"cpu"`
	Labels struct {
		BetaKubernetesIoArch string `json:"beta.kubernetes.io/arch"`
		BetaKubernetesIoOs   string `json:"beta.kubernetes.io/os"`
		IP                   string `json:"ip"`
		KubernetesIoHostname string `json:"kubernetes.io/hostname"`
		Nodetype             string `json:"nodetype"`
	} `json:"labels"`
}
