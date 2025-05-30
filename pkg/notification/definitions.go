package notification

type Function func(map[string]interface{}, []int) (bool, error)
