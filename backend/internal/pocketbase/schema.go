package pocketbase

type schemaBase struct {
	SchemaFlags
	Name string `json:"name"`
	Type string `json:"type"`
}

type SchemaFlags struct {
	System        *bool `json:"system"`
	Required      *bool `json:"required,omitempty"`
	Nullable      *bool `json:"nullable,omitempty"`
	Presentable   *bool `json:"presentable,omitempty"`
	OnMountSelect *bool `json:"onMountSelect,omitempty"`
	ToDelete      *bool `json:"toDelete,omitempty"`
}

func (s schemaBase) S() {}

type schema struct {
	schemaBase
	Options schemaType `json:"options,omitempty"`
}

type schemaType interface {
	Type() string
}

type SchemaOptions interface {
	TextOptions | EmailOptions | FileOptions | EditorOptions | UrlOptions | RelationOptions | NumberOptions | DateTimeOptions | JSONOptions | BoolOptions | SelectOptions
	schemaType
}

func SchemaBuilder[T SchemaOptions](name string, options T, fn ...func(*SchemaFlags)) Schema {
	flags := new(SchemaFlags)
	for _, f := range fn {
		f(flags)
	}
	return schema{
		schemaBase: schemaBase{
			Name:        name,
			Type:        typeOfOptions(options),
			SchemaFlags: *flags,
		},
		Options: options,
	}
}

func typeOfOptions[T SchemaOptions](o T) string {
	return o.Type()
}

type TextOptions struct {
	MinLen int    `json:"minLength,omitempty"`
	MaxLen int    `json:"maxLength,omitempty"`
	Regex  string `json:"pattern,omitempty"`
}

func (TextOptions) Type() string {
	return "text"
}

type EmailOptions struct {
	ExceptDomains []string `json:"exceptDomains"`
	OnlyDomains   []string `json:"onlyDomains"`
}

func (EmailOptions) Type() string {
	return "email"
}

type FileOptions struct {
	AllowedMimeTypes []string `json:"allowedMimeTypes,omitempty"`
	ThumbSizes       []string `json:"thumbs,omitempty"`
	MaxSize          int      `json:"maxSize"`   // required
	MaxSelect        int      `json:"maxSelect"` // required (multiple)
	Protected        *bool    `json:"protected,omitempty"`
}

func (FileOptions) Type() string {
	return "file"
}

type EditorOptions struct {
	StripDomain *bool `json:"convertUrls,omitempty"`
}

func (EditorOptions) Type() string {
	return "editor"
}

type UrlOptions struct {
	ExceptDomains []string `json:"exceptDomains"`
	OnlyDomains   []string `json:"onlyDomains"`
}

func (UrlOptions) Type() string {
	return "url"
}

type RelationOptions struct {
	Collection    string `json:"collectionId"` //required
	CascadeDelete bool   `json:"cascadeDelete"`
	MaxSelect     uint   `json:"maxSelect"`
}

func (RelationOptions) Type() string {
	return "relation"
}

type NumberOptions struct {
	Min        float64 `json:"min,omitempty"`
	Max        float64 `json:"max,omitempty"`
	NoDecimals *bool   `json:"noDecimal,omitempty"`
}

func (NumberOptions) Type() string {
	return "number"
}

type DateTimeOptions struct {
	MinDate string `json:"min,omitempty"`
	MaxDate string `json:"max,omitempty"`
}

func (DateTimeOptions) Type() string {
	return "date"
}

type JSONOptions struct {
	MaxSize int `json:"maxSize"` // required
}

func (JSONOptions) Type() string {
	return "json"
}

type BoolOptions struct {
}

func (BoolOptions) Type() string {
	return "bool"
}

type SelectOptions struct {
	MaxSelect uint     `json:"maxSelect"`
	Values    []string `json:"values"`
}

func (SelectOptions) Type() string {
	return "select"
}
