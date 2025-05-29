package model

type DeployCreateRequest struct {
	Name        string    `json:"name"`
	Namespace   string    `json:"namespace"`
	Image       string    `json:"image"`
	Resources   Resources `json:"resource"`
	Replicas    string    `json:"replicas"`
	Labels      string    `json:"labels"`
	Annotations string    `json:"annotations"`
}

type Resources struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	GPU    string `json:"gpu"`
}

func (p *DeployCreateRequest) Validae() error {
	//_, err := resource.ParseQuantity(p.Cpus)
	//return err
	return nil
}
