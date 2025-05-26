package model

type PodCreateRequest struct {
	Image     string `json:"image"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	//Cpus      string `json:"cpus"`
}

func (p *PodCreateRequest) Validae() error {
	//_, err := resource.ParseQuantity(p.Cpus)
	//return err
	return nil
}
