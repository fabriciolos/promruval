package validator

import (
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/promql/parser"
)

func getExpressionUsedLabels(expr string) ([]string, error) {
	promQl, err := parser.ParseExpr(expr)
	if err != nil {
		return []string{}, fmt.Errorf("failed to parse expression `%s`: %s", expr, err)
	}
	var usedLabels []string
	parser.Inspect(promQl, func(n parser.Node, ns []parser.Node) error {
		switch v := n.(type) {
		case *parser.AggregateExpr:
			usedLabels = append(usedLabels, v.Grouping...)
		case *parser.VectorSelector:
			for _, m := range v.LabelMatchers {
				usedLabels = append(usedLabels, m.Name)
			}
		case *parser.BinaryExpr:
			if v.VectorMatching != nil {
				usedLabels = append(usedLabels, v.VectorMatching.Include...)
				usedLabels = append(usedLabels, v.VectorMatching.MatchingLabels...)
			}
		}
		return nil
	})
	return usedLabels, nil
}

func checkAllLabelValuesUsed(desired []string, match map[string]string) (foundAll bool, errors []error) {
	foundAll = true
	for _, desiredLabel := range desired {
		labelValue := strings.Split(desiredLabel, ":")
		if len(labelValue) == 2 {
			label := labelValue[0]
			value := labelValue[1]
			if match[label] == "" {
				foundAll = false
				errors = append(errors, fmt.Errorf("Vector %s: Missing label %s", match["__name__"], label))
			} else {
				if match[label] != value {
					foundAll = false
					errors = append(errors, fmt.Errorf("Vector %s: Label %s is %s, not %s", match["__name__"], label, match[label], value))
				}
			}
		} else {
			errors = append(errors, fmt.Errorf("Vector %s: Missing label:value in  %s", match["__name__"], desiredLabel))
			foundAll = false
		}
	}
	return foundAll, errors
}

func getExpressionEachLabels(labels []string, expr string) (bool, []error) {
	var errors []error
	promQl, err := parser.ParseExpr(expr)
	if err != nil {
		return false, []error{fmt.Errorf("failed to parse expression `%s`: %s", expr, err)}
	}
	foundAll := true
	parser.Inspect(promQl, func(n parser.Node, ns []parser.Node) error {
		switch v := n.(type) {
		case *parser.VectorSelector:
			matcher := make(map[string]string)
			for _, m := range v.LabelMatchers {
				matcher[m.Name] = m.Value
			}
			foundEach, erro := checkAllLabelValuesUsed(labels, matcher)
			for _, er := range erro {
				errors = append(errors, er)
			}
			if foundEach && foundAll {
				foundAll = true
			} else {
				foundAll = false
			}
		}
		return nil
	})
	return foundAll, errors
}

func getExpressionSelectors(expr string) ([]string, error) {
	promQl, err := parser.ParseExpr(expr)
	if err != nil {
		return []string{}, fmt.Errorf("failed to parse expression `%s`: %s", expr, err)
	}
	var selectors []string
	parser.Inspect(promQl, func(n parser.Node, ns []parser.Node) error {
		switch v := n.(type) {
		case *parser.VectorSelector:
			s := &parser.VectorSelector{Name: v.Name, LabelMatchers: v.LabelMatchers}
			selectors = append(selectors, s.String())
		}
		return nil
	})
	return selectors, nil
}
