package composer

import (
	"strconv"
	"strings"
)

// Condition is copy of Query
type Condition struct {
	Statement string
	Variables []interface{}
}

type Field string

// Name is a helper function to convert FieldStruct to string.
func (f Field) String() string {
	return string(f)
}

// CondGroup is a helper function to create a group of conditions.
// It's Basically wraps the conditions in parentheses.
// (id > ? AND accrual < ?)
func CondGroup(conditions ...Condition) Condition {
	cond := Condition{}
	var builder strings.Builder
	builder.WriteString("(")
	for _, c := range conditions {
		builder.WriteString(c.Statement)
		cond.Variables = append(cond.Variables, c.Variables...)
	}
	builder.WriteString(")")
	cond.Statement = builder.String()
	return cond
}

// In postgres IN
func (f Field) In(values ...any) Condition {
	cond := Condition{}
	var builder strings.Builder
	builder.WriteString(f.String() + " IN (")
	for i, value := range values {
		builder.WriteString("?")
		cond.Variables = append(cond.Variables, value)
		if i != len(values)-1 {
			builder.WriteString(", ")
		}
	}
	builder.WriteString(")")
	cond.Statement += builder.String()
	return cond
}

// NotIn postgres NOT IN
func (f Field) NotIn(values ...any) Condition {
	cond := Condition{}
	var builder strings.Builder
	builder.WriteString(f.String() + " NOT IN (")
	for i, value := range values {
		builder.WriteString("?")
		cond.Variables = append(cond.Variables, value)
		if i != len(values)-1 {
			builder.WriteString(", ")
		}
	}
	builder.WriteString(")")
	cond.Statement += builder.String()
	return cond
}

// Build returns the query and values.
// Example:
//
//	stm, values := Command(Select, UserID, Status, Accrual, UploadedAt).
//		From(OrdersTable).
//		Where(
//			// Read as "(ID > 12 AND Accrual < 1)"
//			CondGroup(ID.BiggerThan(12), And(), Accrual.LowerThan(1)),
//			CondGroup(Or()),
//			CondGroup(Status.EqualTo(models.OrderStatusNew)),
//	).Build()
//	fmt.Printf("Statement: %s\n Variables: %f\n", stm, values)
//
// result will be:
//
//	Statement: SELECT user_id,status,accrual,uploaded_at FROM orders WHERE (user_id > ? AND accrual < ?) OR (status = ?)
//	Variables: [12 1 new]
func (q Condition) Build() (string, []interface{}) {
	q.Statement += ";"
	// count ? in statement
	cnt := strings.Count(q.Statement, "?")
	for i := 1; i <= cnt; i++ {
		q.Statement = strings.Replace(q.Statement, "?", "$"+strconv.Itoa(i), 1)
	}
	return q.Statement, q.Variables
}

// IsNull postgres IS NULL
func (f Field) IsNull() Condition {
	cond := Condition{}
	cond.Statement += f.String() + " IS NULL"
	return cond
}

// Like postgres LIKE
func (f Field) Like(value any) Condition {
	cond := Condition{}
	cond.Statement += f.String() + " LIKE ?"
	cond.Variables = append(cond.Variables, value)
	return cond
}

// And Condition adds string " AND " to the statement.
func And() Condition {
	return Condition{Statement: " AND "}
}

// Or Condition adds string " OR " to the statement.
//func Or() Condition {
//	return Condition{Statement: " OR "}
//}

// BiggerThan postgres >
func (f Field) BiggerThan(value any) Condition {
	cond := Condition{}
	cond.Statement += f.String() + " > ?"
	cond.Variables = append(cond.Variables, value)
	return cond
}

// LowerThan postgres <
func (f Field) LowerThan(value any) Condition {
	cond := Condition{}
	cond.Statement += f.String() + " < ?"
	cond.Variables = append(cond.Variables, value)
	return cond
}

// EqualTo postgres =
func (f Field) EqualTo(value any) Condition {
	cond := Condition{}
	cond.Statement += f.String() + " = ?"
	cond.Variables = append(cond.Variables, value)
	return cond
}

// NotEqualTo postgres !=
func (f Field) NotEqualTo(value any) Condition {
	cond := Condition{}
	cond.Statement += f.String() + " != ?"
	cond.Variables = append(cond.Variables, value)
	return cond
}
