package h

type AttributeR struct {
	Name  string
	Value string
}

type KeyValue[T any] struct {
	Key   string
	Value T
}

type TextContent struct {
	Content string
}

type RawContent struct {
	Content string
}

type ChildList struct {
	Children []Ren
}

func NewAttribute(name string, value string) *AttributeR {
	return &AttributeR{
		Name:  name,
		Value: value,
	}
}

func NewRawContent(content string) *RawContent {
	return &RawContent{
		Content: content,
	}
}

func NewTextContent(content string) *TextContent {
	return &TextContent{
		Content: content,
	}
}

func NewChildList(children ...Ren) *ChildList {
	if len(children) == 0 {
		return &ChildList{Children: make([]Ren, 0)}
	}
	hasNil := false
	for _, c := range children {
		if c == nil {
			hasNil = true
			break
		}
	}
	if !hasNil {
		return &ChildList{Children: children}
	}
	return &ChildList{
		Children: Filter(children, func(item Ren) bool {
			return item != nil
		}),
	}
}
