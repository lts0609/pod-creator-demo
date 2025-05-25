package main

import ()

type PodCreateRequest struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`
	Cpu       string `json:"cpu"`
}
type PodCreateResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`
	Cpu       string `json:"cpu"`
}

func main() {
	gin := gin.Default()

}
