package trainer

var (
	KnownTrainingTypes = []string{"mpijob", "runai"}
	KnownServingTypes  = []string{"tf-serving", "trt-serving", "custom-serving"}
)