package notification

type Function func(map[string]interface{}) (bool, error)
