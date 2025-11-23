package runtime

import (
	"fmt"
	"strings"
)

// Predicate represents a SQL predicate
type Predicate interface {
	String() string
	Args() []interface{}
}

// Query is the interface for query builders
type Query interface {
	Where(predicates ...Predicate) Query
}

// Basic predicates

type eqPredicate struct {
	field string
	value interface{}
}

func (p *eqPredicate) String() string {
	return fmt.Sprintf("%s = $%d", p.field, 1)
}

func (p *eqPredicate) Args() []interface{} {
	return []interface{}{p.value}
}

// EQ creates an equality predicate
func EQ(field string, value interface{}) Predicate {
	return &eqPredicate{field: field, value: value}
}

type neqPredicate struct {
	field string
	value interface{}
}

func (p *neqPredicate) String() string {
	return fmt.Sprintf("%s != $%d", p.field, 1)
}

func (p *neqPredicate) Args() []interface{} {
	return []interface{}{p.value}
}

// NEQ creates a not-equal predicate
func NEQ(field string, value interface{}) Predicate {
	return &neqPredicate{field: field, value: value}
}

type gtPredicate struct {
	field string
	value interface{}
}

func (p *gtPredicate) String() string {
	return fmt.Sprintf("%s > $%d", p.field, 1)
}

func (p *gtPredicate) Args() []interface{} {
	return []interface{}{p.value}
}

// GT creates a greater-than predicate
func GT(field string, value interface{}) Predicate {
	return &gtPredicate{field: field, value: value}
}

type gtePredicate struct {
	field string
	value interface{}
}

func (p *gtePredicate) String() string {
	return fmt.Sprintf("%s >= $%d", p.field, 1)
}

func (p *gtePredicate) Args() []interface{} {
	return []interface{}{p.value}
}

// GTE creates a greater-than-or-equal predicate
func GTE(field string, value interface{}) Predicate {
	return &gtePredicate{field: field, value: value}
}

type ltPredicate struct {
	field string
	value interface{}
}

func (p *ltPredicate) String() string {
	return fmt.Sprintf("%s < $%d", p.field, 1)
}

func (p *ltPredicate) Args() []interface{} {
	return []interface{}{p.value}
}

// LT creates a less-than predicate
func LT(field string, value interface{}) Predicate {
	return &ltPredicate{field: field, value: value}
}

type ltePredicate struct {
	field string
	value interface{}
}

func (p *ltePredicate) String() string {
	return fmt.Sprintf("%s <= $%d", p.field, 1)
}

func (p *ltePredicate) Args() []interface{} {
	return []interface{}{p.value}
}

// LTE creates a less-than-or-equal predicate
func LTE(field string, value interface{}) Predicate {
	return &ltePredicate{field: field, value: value}
}

type inPredicate struct {
	field  string
	values []interface{}
}

func (p *inPredicate) String() string {
	placeholders := make([]string, len(p.values))
	for i := range p.values {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf("%s IN (%s)", p.field, strings.Join(placeholders, ", "))
}

func (p *inPredicate) Args() []interface{} {
	return p.values
}

// In creates an IN predicate
func In(field string, values ...interface{}) Predicate {
	return &inPredicate{field: field, values: values}
}

type notInPredicate struct {
	field  string
	values []interface{}
}

func (p *notInPredicate) String() string {
	placeholders := make([]string, len(p.values))
	for i := range p.values {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf("%s NOT IN (%s)", p.field, strings.Join(placeholders, ", "))
}

func (p *notInPredicate) Args() []interface{} {
	return p.values
}

// NotIn creates a NOT IN predicate
func NotIn(field string, values ...interface{}) Predicate {
	return &notInPredicate{field: field, values: values}
}

type isNullPredicate struct {
	field string
}

func (p *isNullPredicate) String() string {
	return fmt.Sprintf("%s IS NULL", p.field)
}

func (p *isNullPredicate) Args() []interface{} {
	return nil
}

// IsNull creates an IS NULL predicate
func IsNull(field string) Predicate {
	return &isNullPredicate{field: field}
}

type notNullPredicate struct {
	field string
}

func (p *notNullPredicate) String() string {
	return fmt.Sprintf("%s IS NOT NULL", p.field)
}

func (p *notNullPredicate) Args() []interface{} {
	return nil
}

// NotNull creates an IS NOT NULL predicate
func NotNull(field string) Predicate {
	return &notNullPredicate{field: field}
}

type containsPredicate struct {
	field string
	value string
}

func (p *containsPredicate) String() string {
	return fmt.Sprintf("%s LIKE $%d", p.field, 1)
}

func (p *containsPredicate) Args() []interface{} {
	return []interface{}{"%" + p.value + "%"}
}

// Contains creates a LIKE predicate for substring match
func Contains(field, value string) Predicate {
	return &containsPredicate{field: field, value: value}
}

type hasPrefixPredicate struct {
	field string
	value string
}

func (p *hasPrefixPredicate) String() string {
	return fmt.Sprintf("%s LIKE $%d", p.field, 1)
}

func (p *hasPrefixPredicate) Args() []interface{} {
	return []interface{}{p.value + "%"}
}

// HasPrefix creates a LIKE predicate for prefix match
func HasPrefix(field, value string) Predicate {
	return &hasPrefixPredicate{field: field, value: value}
}

type hasSuffixPredicate struct {
	field string
	value string
}

func (p *hasSuffixPredicate) String() string {
	return fmt.Sprintf("%s LIKE $%d", p.field, 1)
}

func (p *hasSuffixPredicate) Args() []interface{} {
	return []interface{}{"%" + p.value}
}

// HasSuffix creates a LIKE predicate for suffix match
func HasSuffix(field, value string) Predicate {
	return &hasSuffixPredicate{field: field, value: value}
}

// Relationship predicates

type hasEdgePredicate struct {
	edge string
}

func (p *hasEdgePredicate) String() string {
	return fmt.Sprintf("EXISTS (SELECT 1 FROM %s)", p.edge)
}

func (p *hasEdgePredicate) Args() []interface{} {
	return nil
}

// HasEdge creates a predicate to check if a relationship exists
func HasEdge(edge string) Predicate {
	return &hasEdgePredicate{edge: edge}
}

type hasEdgeWithPredicate struct {
	edge       string
	predicates []Predicate
}

func (p *hasEdgeWithPredicate) String() string {
	// TODO: Implement proper subquery
	return fmt.Sprintf("EXISTS (SELECT 1 FROM %s WHERE ...)", p.edge)
}

func (p *hasEdgeWithPredicate) Args() []interface{} {
	var args []interface{}
	for _, pred := range p.predicates {
		args = append(args, pred.Args()...)
	}
	return args
}

// HasEdgeWith creates a predicate to filter by relationship
func HasEdgeWith(edge string, predicates ...Predicate) Predicate {
	return &hasEdgeWithPredicate{edge: edge, predicates: predicates}
}

// Logical operators

type andPredicate struct {
	predicates []Predicate
}

func (p *andPredicate) String() string {
	var parts []string
	for _, pred := range p.predicates {
		parts = append(parts, pred.String())
	}
	return "(" + strings.Join(parts, " AND ") + ")"
}

func (p *andPredicate) Args() []interface{} {
	var args []interface{}
	for _, pred := range p.predicates {
		args = append(args, pred.Args()...)
	}
	return args
}

// And combines predicates with AND
func And(predicates ...Predicate) Predicate {
	return &andPredicate{predicates: predicates}
}

type orPredicate struct {
	predicates []Predicate
}

func (p *orPredicate) String() string {
	var parts []string
	for _, pred := range p.predicates {
		parts = append(parts, pred.String())
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func (p *orPredicate) Args() []interface{} {
	var args []interface{}
	for _, pred := range p.predicates {
		args = append(args, pred.Args()...)
	}
	return args
}

// Or combines predicates with OR
func Or(predicates ...Predicate) Predicate {
	return &orPredicate{predicates: predicates}
}

type notPredicate struct {
	predicate Predicate
}

func (p *notPredicate) String() string {
	return "NOT (" + p.predicate.String() + ")"
}

func (p *notPredicate) Args() []interface{} {
	return p.predicate.Args()
}

// Not negates a predicate
func Not(predicate Predicate) Predicate {
	return &notPredicate{predicate: predicate}
}
