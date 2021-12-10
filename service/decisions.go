package service

import (
	"dang/model"
	"dang/repository"
	"dang/view"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type DecisionService struct {
	repository repository.Repository
}

func NewDecisionService(repository repository.Repository) DecisionService {
	return DecisionService{repository}
}

func (ds DecisionService) SaveDecisions(saveDecisionRequest view.SaveDecisionRequest) error {
	decisionRequests := saveDecisionRequest.Decisions

	for _, decisionRequest := range decisionRequests {
		inputFormFields, outputFormFields := []model.FormField{}, []model.FormField{}
		for _, field := range decisionRequest.InputForm {
			var decisionChainLink *model.DecisionChainLink
			var err error
			if len(field.Value) > 0 {
				decisionChainLink, err = createChainLink(field, field.Key)
				if err != nil {
					return err
				}
			}

			inputFormFields = append(inputFormFields, model.FormField{
				Key:         field.Key,
				Title:       field.Title,
				Description: field.Description,
				Unit:        field.Unit,
				Type:        field.Type,
				ChainLink:   decisionChainLink,
			})
		}
		for _, field := range decisionRequest.OutputForm {
			rule, err := createRule(field.Key, field.Type, field.Rule)
			if err != nil {
				return err
			}

			outputFormFields = append(outputFormFields, model.FormField{
				Key:         field.Key,
				Title:       field.Title,
				Description: field.Description,
				Unit:        field.Unit,
				Type:        field.Type,
				Rule:        &rule,
			})
		}

		decision := model.Decision{
			Name:       decisionRequest.Name,
			InputForm:  inputFormFields,
			OutputForm: outputFormFields,
		}

		err := ds.repository.InsertDecision(decision)
		if err != nil {
			return err
		}
	}

	return nil
}

func createChainLink(request view.FormFieldRequest, formKey string) (*model.DecisionChainLink, error) {
	sourceComponents := strings.Split(request.Value, ".")
	if sourceComponents[0] != "Decision" {
		return nil, errors.New("unrecognized source component")
	}

	return &model.DecisionChainLink{
		DecisionName:   sourceComponents[1],
		SourceKey:      sourceComponents[2],
		DestinationKey: formKey,
	}, nil
}

func createRule(formKey, formValueType string, request view.RuleRequest) (string, error) {
	switch request.Type {
	case "IntervalMap":
		if len(request.Args[0]) <= 0 {
			return "", errors.New("invalid args")
		}
		arg := request.Args[0]
		intervalMap, ok := request.Rule.(map[string]interface{})
		if !ok {
			return "", errors.New("invalid rule")
		}

		return createIntervalMapRule(formKey, formValueType, arg, intervalMap)
	case "BoolMap":
		if len(request.Args[0]) <= 0 {
			return "", errors.New("invalid args")
		}
		arg := request.Args[0]
		boolMap, ok := request.Rule.(map[string]interface{})
		if !ok {
			return "", errors.New("invalid rule")
		}

		return createBoolMapRule(formKey, formValueType, arg, boolMap), nil
	case "ThresholdMap":
		if len(request.Args[0]) <= 0 {
			return "", errors.New("invalid args")
		}
		arg := request.Args[0]
		thresholdMap, ok := request.Rule.(map[string]interface{})
		if !ok {
			return "", errors.New("invalid rule")
		}

		return createThresholdMapRule(formKey, formValueType, arg, thresholdMap)
	case "Func":
		funcName, ok := request.Rule.(string)
		if !ok {
			return "", errors.New("invalid rule")
		}

		return createFuncRule(formKey, funcName, request.Args), nil
	default:
		return "unsupported rule type", nil
	}
}

func createFuncRule(formKey, funcName string, args []string) string {
	return fmt.Sprintf(`{"%s":{{%s .%s}}}`, formKey, funcName, strings.Join(args, " ."))
}

func createIntervalMapRule(formKey, formValueType, arg string, intervalMap map[string]interface{}) (string, error) {
	validatedIntervals, err := getValidatedIntervals(intervalMap)
	if err != nil {
		return "", err
	}

	rule := ""
    for _, interval := range validatedIntervals {
    	templateCondition, err := intervalToTemplateCondition(interval, arg)
    	if err != nil {
    		return "", err
    	}
    	rule = rule + templateCondition + fmt.Sprintf(" %v ", createOutputFormFieldValue(intervalMap[interval], formValueType))
    }
    rule = rule + `{{end}}`
	return fmt.Sprintf(`{"%s": %s}`, formKey, rule), nil
}

func createBoolMapRule(formKey, formValueType, arg string, boolMap map[string]interface{}) string {
	return fmt.Sprintf(`{"%s":{{if .%s}}%v{{else}}%v{{end}}}`,
		formKey,
		arg,
		createOutputFormFieldValue(boolMap["true"], formValueType),
		createOutputFormFieldValue(boolMap["false"], formValueType),
	)
}

func createThresholdMapRule(formKey, formValueType, arg string, thresholdMap map[string]interface{}) (string, error) {
	threshold, trueCondition, falseCondition, comparisonOperator, err := getValidatedThreshold(thresholdMap)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`{"%s":{{if %s .%s %v}}%v{{else}}%v{{end}}}`,
		formKey, comparisonOperator, arg, threshold,
		createOutputFormFieldValue(thresholdMap[trueCondition], formValueType),
		createOutputFormFieldValue(thresholdMap[falseCondition], formValueType),
	), nil
}

func createOutputFormFieldValue(value interface{}, valueType string) interface{} {
	switch valueType {
	case "Number", "Bool":
		return value
	case "String":
		return fmt.Sprintf(`"%v"`, value)
	default:
		return value
	}
}

func intervalToTemplateCondition(interval, operandKey string) (string, error) {
	if strings.HasPrefix(interval, "(...") {
		return bottomIntervalToTemplate(strings.TrimPrefix(interval, "(..."), operandKey)
	}

	if strings.HasSuffix(interval, "...)") {
		return `{{else}}`, nil
	}

	bounds := strings.Split(interval, "...")
	lowerBoundStr, upperBoundStr := bounds[0], bounds[1]

	lowerBound, lowerBoundComparisonOperator, err := getBound(lowerBoundStr)
	if err != nil {
		return "", err
	}

	upperBound, upperBoundComparisonOperator, err := getBound(upperBoundStr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`{{else if (and (%s .%s %v) (%s .%s %v))}}`,
		lowerBoundComparisonOperator, operandKey, lowerBound,
		upperBoundComparisonOperator, operandKey, upperBound), nil
}

func getBound(boundStr string) (bound float64, comparisonOperator string, err error) {
	switch {
	case strings.HasPrefix(boundStr, "["):
		comparisonOperator = "greater_than_or_equals"
		bound, err = strconv.ParseFloat(strings.TrimPrefix(boundStr, "["), 64)
		return
	case strings.HasPrefix(boundStr, "("):
		comparisonOperator = "greater_than"
		bound, err = strconv.ParseFloat(strings.TrimPrefix(boundStr, "("), 64)
		return
	case strings.HasSuffix(boundStr, "]"):
		comparisonOperator = "less_than_or_equals"
		bound, err = strconv.ParseFloat(strings.TrimSuffix(boundStr, "]"), 64)
		return
	case strings.HasSuffix(boundStr, ")"):
		comparisonOperator = "less_than"
		bound, err = strconv.ParseFloat(strings.TrimSuffix(boundStr, ")"), 64)
		return
	default:
		err = errors.New("invalid bound")
		return
	}
}

func bottomIntervalToTemplate(bound, operandKey string) (string, error) {
	if strings.HasSuffix(bound, "]") {
		boundVal, err := strconv.ParseFloat(strings.TrimSuffix(bound, "]"), 64)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`{{if less_than_or_equals .%s %v}}`, operandKey, boundVal), nil
	}

	if strings.HasSuffix(bound, ")") {
		boundVal, err := strconv.ParseFloat(strings.TrimSuffix(bound, ")"), 64)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`{{if less_than .%s %v}}`, operandKey, boundVal), nil
	}

	return "", errors.New("invalid interval")
}

func getValidatedIntervals(intervalMap map[string]interface{}) (validatedIntervals []string, err error) {
	intervals, iteratedIntervals := []string{}, []string{}
	for k, _ := range intervalMap {
		intervals = append(intervals, k)
		iteratedIntervals = append(iteratedIntervals, k)
	}

	if len(intervals) == 1 {
		err = errors.New("invalid intervals")
		return
	}

	lastIterated := ""
	lastIteratedIndex := 0
	iteratingPrefix := "(..."
	final := false
	for {
		var possibleIndices []int
		for index, interval := range iteratedIntervals {
			if !strings.HasPrefix(interval, iteratingPrefix) {
				continue
			}

			possibleIndices = append(possibleIndices, index)
		}
		if len(possibleIndices) < 1 {
			err = errors.New("non-exhaustive intervals")
			return
		}
		if len(possibleIndices) > 1 {
			err = errors.New("overlapping intervals")
			return
		}
		lastIteratedIndex = possibleIndices[0]
		lastIterated = iteratedIntervals[lastIteratedIndex]

		var upperBoundary float64
		switch {
		case strings.HasSuffix(lastIterated, "...)"):
			if iteratingPrefix == "(..." {
				err = errors.New("invalid interval")
				return
			}
			final = true
			break
		case strings.HasSuffix(lastIterated, "]"):
			lastIteratedTrimmed := strings.TrimPrefix(lastIterated, iteratingPrefix)
			upperBoundary, err = strconv.ParseFloat(strings.TrimSuffix(lastIteratedTrimmed, "]"), 64)
			if err != nil {
				return
			}
			iteratingPrefix = fmt.Sprintf("(%v...", upperBoundary)
		case strings.HasSuffix(lastIterated, ")"):
			lastIteratedTrimmed := strings.TrimPrefix(lastIterated, iteratingPrefix)
			upperBoundary, err = strconv.ParseFloat(strings.TrimSuffix(lastIteratedTrimmed, ")"), 64)
			if err != nil {
				return
			}
			iteratingPrefix = fmt.Sprintf("[%v...", upperBoundary)
		default:
			err = errors.New("invalid interval notation")
			return
		}

		validatedIntervals = append(validatedIntervals, lastIterated)
		if final {
			break
		}

		iteratedIntervals = append(iteratedIntervals[:lastIteratedIndex], iteratedIntervals[lastIteratedIndex+1:]...)
	}

	if len(validatedIntervals) != len(intervals) {
		err = errors.New("overlapping intervals")
		return
	}

	return
}

func getValidatedThreshold(thresholdMap map[string]interface{}) (threshold float64, trueCondition, falseCondition, templateComparisonOperator string, err error) {
	conditions := []string{}
	for k, _ := range thresholdMap {
		conditions = append(conditions, k)
	}
	if len(conditions) != 2 {
		err = errors.New("invalid threshold condition map")
		return
	}

	comparisonOperator := ""
	trueCondition, falseCondition = conditions[0], conditions[1]
	switch {
	case strings.HasPrefix(trueCondition, ">") && strings.HasPrefix(falseCondition, "<="):
		comparisonOperator = ">"
		templateComparisonOperator = "greater_than"
	case strings.HasPrefix(trueCondition, ">=") && strings.HasPrefix(falseCondition, "<"):
		comparisonOperator = ">="
		templateComparisonOperator = "greater_than_or_equals"
	case strings.HasPrefix(trueCondition, "<") && strings.HasPrefix(falseCondition, ">="):
		comparisonOperator = "<"
		templateComparisonOperator = "less_than"
	case strings.HasPrefix(trueCondition, "<=") && strings.HasPrefix(falseCondition, ">"):
		comparisonOperator = "<="
		templateComparisonOperator = "less_than_or_equals"
	default:
		err = errors.New("invalid threshold map")
		return
	}

	threshold, err = strconv.ParseFloat(strings.TrimPrefix(trueCondition, comparisonOperator), 64)
	return
}