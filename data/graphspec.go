package data

type GraphSpec struct {
	Fields []Field
}

type Field struct {
	ID      string
	Name    string
	Counter bool
	Marker  bool
}
