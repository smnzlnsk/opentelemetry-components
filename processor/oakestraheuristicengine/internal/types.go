package internal

type EvaluationRequestElement struct {
	ServiceIdentifier string
	Params            map[string]interface{}
}

type EvaluationRequest struct {
	Elements []EvaluationRequestElement
}

type EvaluationResultElement struct {
	ServiceIdentifier string
	Result            float64
	Error             error
}

type EvaluationResult struct {
	Elements []EvaluationResultElement
}
