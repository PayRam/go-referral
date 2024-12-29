package db

type QueryCondition struct {
	Field    string      // Field name
	Operator string      // Operator (e.g., "=", "<", ">", "LIKE")
	Value    interface{} // Value to compare against
}
