package hack

/*
We are using generated template.go for serialized kubernetes and helm assets
*/

//go:generate go-bindata -o ../pkg/embeded/charts/charts_generated.go -pkg charts -prefix ../charts/ ../charts/...
//go:generate go-bindata -o ../pkg/embeded/stage/stage_generated.go -pkg stage -prefix ../stage/ ../stage/...
