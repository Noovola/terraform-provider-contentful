package contenttype

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/elliotchance/pie/v2"
	"github.com/flaconi/contentful-go"
	"github.com/flaconi/contentful-go/pkgs/model"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

// ContentType is the main resource schema data
type ContentType struct {
	ID                  types.String `tfsdk:"id"`
	SpaceId             types.String `tfsdk:"space_id"`
	Environment         types.String `tfsdk:"environment"`
	Name                types.String `tfsdk:"name"`
	DisplayField        types.String `tfsdk:"display_field"`
	Description         types.String `tfsdk:"description"`
	Version             types.Int64  `tfsdk:"version"`
	VersionControls     types.Int64  `tfsdk:"version_controls"`
	Fields              []Field      `tfsdk:"fields"`
	ManageFieldControls types.Bool   `tfsdk:"manage_field_controls"`
	Sidebar             []Sidebar    `tfsdk:"sidebar"`
}

type Sidebar struct {
	WidgetId        types.String         `tfsdk:"widget_id"`
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	Disabled        types.Bool           `tfsdk:"disabled"`
}

type Field struct {
	Id           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	Type         types.String  `tfsdk:"type"`
	LinkType     types.String  `tfsdk:"link_type"`
	Required     types.Bool    `tfsdk:"required"`
	Localized    types.Bool    `tfsdk:"localized"`
	Disabled     types.Bool    `tfsdk:"disabled"`
	Omitted      types.Bool    `tfsdk:"omitted"`
	Validations  []Validation  `tfsdk:"validations"`
	Items        *Items        `tfsdk:"items"`
	Control      *Control      `tfsdk:"control"`
	DefaultValue *DefaultValue `tfsdk:"default_value"`
}

type DefaultValue struct {
	Bool   types.Map `tfsdk:"bool"`
	String types.Map `tfsdk:"string"`
}

func (d *DefaultValue) Draft() map[string]any {
	var defaultValues = map[string]any{}

	if !d.String.IsNull() && !d.String.IsUnknown() {

		for k, v := range d.String.Elements() {
			defaultValues[k] = v.(types.String).ValueString()
		}
	}

	if !d.Bool.IsNull() && !d.Bool.IsUnknown() {

		for k, v := range d.Bool.Elements() {
			defaultValues[k] = v.(types.Bool).ValueBool()
		}
	}

	return defaultValues
}

type Control struct {
	WidgetId        types.String `tfsdk:"widget_id"`
	WidgetNamespace types.String `tfsdk:"widget_namespace"`
	Settings        *Settings    `tfsdk:"settings"`
}

type Validation struct {
	Unique            types.Bool     `tfsdk:"unique"`
	Size              *Size          `tfsdk:"size"`
	Range             *Size          `tfsdk:"range"`
	AssetFileSize     *Size          `tfsdk:"asset_file_size"`
	Regexp            *Regexp        `tfsdk:"regexp"`
	LinkContentType   []types.String `tfsdk:"link_content_type"`
	LinkMimetypeGroup []types.String `tfsdk:"link_mimetype_group"`
	In                []types.String `tfsdk:"in"`
	EnabledMarks      []types.String `tfsdk:"enabled_marks"`
	EnabledNodeTypes  []types.String `tfsdk:"enabled_node_types"`
	Message           types.String   `tfsdk:"message"`
}

func (v Validation) Draft() model.FieldValidation {

	if !v.Unique.IsUnknown() && !v.Unique.IsNull() {
		return model.FieldValidationUnique{
			Unique: v.Unique.ValueBool(),
		}
	}

	if v.Size != nil {
		return model.FieldValidationSize{
			Size: &model.MinMax{
				Min: v.Size.Min.ValueFloat64Pointer(),
				Max: v.Size.Max.ValueFloat64Pointer(),
			},
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	if v.Range != nil {
		return model.FieldValidationRange{
			Range: &model.MinMax{
				Min: v.Range.Min.ValueFloat64Pointer(),
				Max: v.Range.Max.ValueFloat64Pointer(),
			},
			ErrorMessage: v.Message.ValueString(),
		}
	}

	if v.AssetFileSize != nil {
		return model.FieldValidationFileSize{
			Size: &model.MinMax{
				Min: v.AssetFileSize.Min.ValueFloat64Pointer(),
				Max: v.AssetFileSize.Max.ValueFloat64Pointer(),
			},
		}
	}

	if v.Regexp != nil {
		return model.FieldValidationRegex{
			Regex: &model.Regex{
				Pattern: v.Regexp.Pattern.ValueString(),
			},
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	if len(v.LinkContentType) > 0 {
		return model.FieldValidationLink{
			LinkContentType: pie.Map(v.LinkContentType, func(t types.String) string {
				return t.ValueString()
			}),
		}
	}

	if len(v.LinkMimetypeGroup) > 0 {
		return model.FieldValidationMimeType{
			MimeTypes: pie.Map(v.LinkMimetypeGroup, func(t types.String) string {
				return t.ValueString()
			}),
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	if len(v.In) > 0 {
		return model.FieldValidationPredefinedValues{
			In: pie.Map(v.In, func(t types.String) any {
				return t.ValueString()
			}),
		}
	}

	if len(v.EnabledMarks) > 0 {
		return model.FieldValidationEnabledMarks{
			Marks: pie.Map(v.EnabledMarks, func(t types.String) string {
				return t.ValueString()
			}),
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	if len(v.EnabledNodeTypes) > 0 {
		return model.FieldValidationEnabledNodeTypes{
			NodeTypes: pie.Map(v.EnabledNodeTypes, func(t types.String) string {
				return t.ValueString()
			}),
			ErrorMessage: v.Message.ValueStringPointer(),
		}
	}

	return nil
}

type Size struct {
	Min types.Float64 `tfsdk:"min"`
	Max types.Float64 `tfsdk:"max"`
}

type Regexp struct {
	Pattern types.String `tfsdk:"pattern"`
}

func (f *Field) Equal(n *model.Field) bool {

	if n.Type != f.Type.ValueString() {
		return false
	}

	if n.ID != f.Id.ValueString() {
		return false
	}

	if n.Name != f.Name.ValueString() {
		return false
	}

	if n.LinkType != f.LinkType.ValueString() {
		return false
	}

	if n.Required != f.Required.ValueBool() {
		return false
	}

	if n.Omitted != f.Omitted.ValueBool() {
		return false
	}

	if n.Disabled != f.Disabled.ValueBool() {
		return false
	}

	if n.Localized != f.Localized.ValueBool() {
		return false
	}

	if f.Items == nil && n.Items != nil {
		return false
	}

	if f.Items != nil && !f.Items.Equal(n.Items) {
		return false
	}

	if len(f.Validations) != len(n.Validations) {
		return false
	}

	for idx, validation := range pie.Map(f.Validations, func(t Validation) model.FieldValidation {
		return t.Draft()
	}) {
		cfVal := n.Validations[idx]

		if !reflect.DeepEqual(validation, cfVal) {
			return false
		}

	}

	if f.DefaultValue != nil && !reflect.DeepEqual(f.DefaultValue.Draft(), n.DefaultValue) {
		return false
	}

	return true
}

func (f *Field) ToNative() (*model.Field, error) {

	contentfulField := &model.Field{
		ID:        f.Id.ValueString(),
		Name:      f.Name.ValueString(),
		Type:      f.Type.ValueString(),
		Localized: f.Localized.ValueBool(),
		Required:  f.Required.ValueBool(),
		Disabled:  f.Disabled.ValueBool(),
		Omitted:   f.Omitted.ValueBool(),
		Validations: pie.Map(f.Validations, func(t Validation) model.FieldValidation {
			return t.Draft()
		}),
	}

	if !f.LinkType.IsNull() && !f.LinkType.IsUnknown() {
		contentfulField.LinkType = f.LinkType.ValueString()
	}

	if contentfulField.Type == model.FieldTypeArray {
		items, errItem := f.Items.ToNative()

		if errItem != nil {
			return nil, errItem
		}

		contentfulField.Items = items
	}

	if f.DefaultValue != nil {
		contentfulField.DefaultValue = f.DefaultValue.Draft()
	}

	return contentfulField, nil
}

func getTypeOfMap(mapValues map[string]any) (*string, error) {
	for _, v := range mapValues {
		switch c := v.(type) {
		case string:
			t := "string"
			return &t, nil
		case bool:
			t := "bool"
			return &t, nil
		default:
			return nil, fmt.Errorf("The default type %T is not supported by the provider", c)
		}
	}

	return nil, nil
}

func (f *Field) Import(n *model.Field, c []contentful.Controls) error {
	f.Id = types.StringValue(n.ID)
	f.Name = types.StringValue(n.Name)
	f.Type = types.StringValue(n.Type)
	f.LinkType = utils.FromOptionalString(n.LinkType)
	f.Required = types.BoolValue(n.Required)
	f.Omitted = types.BoolValue(n.Omitted)
	f.Localized = types.BoolValue(n.Localized)
	f.Disabled = types.BoolValue(n.Disabled)

	defaultValueType, err := getTypeOfMap(n.DefaultValue)

	if err != nil {
		return err
	}

	if defaultValueType != nil {

		f.DefaultValue = &DefaultValue{
			Bool:   types.MapNull(types.BoolType),
			String: types.MapNull(types.StringType),
		}

		switch *defaultValueType {
		case "string":
			stringMap := map[string]attr.Value{}

			for k, v := range n.DefaultValue {
				stringMap[k] = types.StringValue(v.(string))
			}

			f.DefaultValue.String = types.MapValueMust(types.StringType, stringMap)
		case "bool":
			boolMap := map[string]attr.Value{}

			for k, v := range n.DefaultValue {
				boolMap[k] = types.BoolValue(v.(bool))
			}

			f.DefaultValue.Bool = types.MapValueMust(types.BoolType, boolMap)
		}

	}

	validations, err := getValidations(n.Validations)

	if err != nil {
		return err
	}

	f.Validations = validations

	if n.Type == model.FieldTypeArray {

		itemValidations, err := getValidations(n.Items.Validations)

		if err != nil {
			return err
		}

		f.Items = &Items{
			Type:        types.StringValue(n.Items.Type),
			LinkType:    types.StringPointerValue(n.Items.LinkType),
			Validations: itemValidations,
		}
	}

	idx := pie.FindFirstUsing(c, func(control contentful.Controls) bool {
		return n.ID == control.FieldID
	})

	if idx != -1 && c[idx].WidgetID != nil {

		var settings *Settings

		if c[idx].Settings != nil {
			settings = &Settings{}

			settings.Import(c[idx].Settings)
		}

		f.Control = &Control{
			WidgetId:        types.StringPointerValue(c[idx].WidgetID),
			WidgetNamespace: types.StringPointerValue(c[idx].WidgetNameSpace),
			Settings:        settings,
		}
	}

	return nil
}

type Settings struct {
	HelpText        types.String `tfsdk:"help_text"`
	TrueLabel       types.String `tfsdk:"true_label"`
	FalseLabel      types.String `tfsdk:"false_label"`
	Stars           types.Int64  `tfsdk:"stars"`
	Format          types.String `tfsdk:"format"`
	TimeFormat      types.String `tfsdk:"ampm"`
	BulkEditing     types.Bool   `tfsdk:"bulk_editing"`
	TrackingFieldId types.String `tfsdk:"tracking_field_id"`
}

func (s *Settings) Import(settings *contentful.Settings) {
	s.HelpText = types.StringPointerValue(settings.HelpText)
	s.TrueLabel = types.StringPointerValue(settings.TrueLabel)
	s.FalseLabel = types.StringPointerValue(settings.FalseLabel)
	s.Stars = types.Int64PointerValue(settings.Stars)
	s.Format = types.StringPointerValue(settings.Format)
	s.TimeFormat = types.StringPointerValue(settings.AMPM)
	s.BulkEditing = types.BoolPointerValue(settings.BulkEditing)
	s.TrackingFieldId = types.StringPointerValue(settings.TrackingFieldId)
}

func (s *Settings) Draft() *contentful.Settings {
	settings := &contentful.Settings{}

	settings.HelpText = s.HelpText.ValueStringPointer()
	settings.TrueLabel = s.TrueLabel.ValueStringPointer()
	settings.FalseLabel = s.FalseLabel.ValueStringPointer()
	settings.Stars = s.Stars.ValueInt64Pointer()
	settings.Format = s.Format.ValueStringPointer()
	settings.AMPM = s.TimeFormat.ValueStringPointer()
	settings.BulkEditing = s.BulkEditing.ValueBoolPointer()
	settings.TrackingFieldId = s.TrackingFieldId.ValueStringPointer()
	return settings
}

type Items struct {
	Type        types.String `tfsdk:"type"`
	LinkType    types.String `tfsdk:"link_type"`
	Validations []Validation `tfsdk:"validations"`
}

func (i *Items) ToNative() (*model.FieldTypeArrayItem, error) {

	return &model.FieldTypeArrayItem{
		Type: i.Type.ValueString(),
		Validations: pie.Map(i.Validations, func(t Validation) model.FieldValidation {
			return t.Draft()
		}),
		LinkType: i.LinkType.ValueStringPointer(),
	}, nil
}

func (i *Items) Equal(n *model.FieldTypeArrayItem) bool {

	if n == nil {
		return false
	}

	if i.Type.ValueString() != n.Type {
		return false
	}

	if !utils.CompareStringPointer(i.LinkType, n.LinkType) {
		return false
	}

	if len(i.Validations) != len(n.Validations) {
		return false
	}

	for idx, validation := range pie.Map(i.Validations, func(t Validation) model.FieldValidation {
		return t.Draft()
	}) {
		cfVal := n.Validations[idx]

		if !reflect.DeepEqual(validation, cfVal) {
			return false
		}

	}

	return true
}

func (c *ContentType) Draft() (*model.ContentType, error) {

	var fields []*model.Field

	for _, field := range c.Fields {

		nativeField, err := field.ToNative()
		if err != nil {
			return nil, err
		}

		fields = append(fields, nativeField)
	}

	contentfulType := &model.ContentType{
		Name:         c.Name.ValueString(),
		DisplayField: c.DisplayField.ValueString(),
		Fields:       fields,
	}

	if !c.ID.IsUnknown() || !c.ID.IsNull() {
		contentfulType.Sys = &model.EnvironmentSys{
			SpaceSys: model.SpaceSys{
				CreatedSys: model.CreatedSys{
					BaseSys: model.BaseSys{
						ID: c.ID.ValueString(),
					},
				},
			},
		}
	}

	if !c.Description.IsNull() && !c.Description.IsUnknown() {
		contentfulType.Description = c.Description.ValueStringPointer()
	}

	return contentfulType, nil

}

func (c *ContentType) Import(n *model.ContentType, e *contentful.EditorInterface) error {
	c.ID = types.StringValue(n.Sys.ID)
	c.Version = types.Int64Value(int64(n.Sys.Version))

	c.Description = types.StringPointerValue(n.Description)

	c.Name = types.StringValue(n.Name)
	c.DisplayField = types.StringValue(n.DisplayField)

	var fields []Field

	var controls []contentful.Controls
	var sidebar []contentful.Sidebar
	c.VersionControls = types.Int64Value(0)
	if e != nil {
		controls = e.Controls
		sidebar = e.SideBar
		c.VersionControls = types.Int64Value(int64(e.Sys.Version))
	}

	for _, nf := range n.Fields {
		field := &Field{}
		err := field.Import(nf, controls)
		if err != nil {
			return err
		}
		fields = append(fields, *field)
	}

	c.Sidebar = pie.Map(sidebar, func(t contentful.Sidebar) Sidebar {

		settings := jsontypes.NewNormalizedValue("{}")

		if t.Settings != nil {
			data, _ := json.Marshal(t.Settings)
			settings = jsontypes.NewNormalizedValue(string(data))
		}
		return Sidebar{
			WidgetId:        types.StringValue(t.WidgetID),
			WidgetNamespace: types.StringValue(t.WidgetNameSpace),
			Settings:        settings,
			Disabled:        types.BoolValue(t.Disabled),
		}
	})

	c.Fields = fields

	return nil

}

func (c *ContentType) Equal(n *model.ContentType) bool {

	if !utils.CompareStringPointer(c.Description, n.Description) {
		return false
	}

	if c.Name.ValueString() != n.Name {
		return false
	}

	if c.DisplayField.ValueString() != n.DisplayField {
		return false
	}

	if len(c.Fields) != len(n.Fields) {
		return false
	}

	for idxOrg, field := range c.Fields {
		idx := pie.FindFirstUsing(n.Fields, func(f *model.Field) bool {
			return f.ID == field.Id.ValueString()
		})

		if idx == -1 {
			return false
		}

		if !field.Equal(n.Fields[idx]) {
			return false
		}

		// field was moved, it is the same as before but different position
		if idxOrg != idx {
			return false
		}
	}

	return true
}

func (c *ContentType) EqualEditorInterface(n *contentful.EditorInterface) bool {

	if len(c.Fields) != len(n.Controls) {
		return false
	}

	filteredControls := pie.Filter(n.Controls, func(c contentful.Controls) bool {
		return c.WidgetID != nil || c.WidgetNameSpace != nil || c.Settings != nil
	})

	filteredFields := pie.Filter(c.Fields, func(f Field) bool {
		return f.Control != nil
	})

	if len(filteredControls) != len(filteredFields) {
		return false
	}

	for _, field := range filteredFields {
		idx := pie.FindFirstUsing(filteredControls, func(t contentful.Controls) bool {
			return t.FieldID == field.Id.ValueString()
		})

		if idx == -1 {
			return false
		}
		control := filteredControls[idx]

		if field.Control.WidgetId.ValueString() != *control.WidgetID {
			return false
		}

		if field.Control.WidgetNamespace.ValueString() != *control.WidgetNameSpace {
			return false
		}

		if field.Control.Settings == nil && control.Settings != nil {
			return false
		}

		if field.Control.Settings != nil && !reflect.DeepEqual(field.Control.Settings.Draft(), control.Settings) {
			return false
		}
	}

	if len(c.Sidebar) != len(n.SideBar) {
		return false
	}

	for idxOrg, s := range c.Sidebar {
		idx := pie.FindFirstUsing(n.SideBar, func(t contentful.Sidebar) bool {
			return t.WidgetID == s.WidgetId.ValueString()
		})

		if idx == -1 {
			return false
		}

		// field was moved, it is the same as before but different position
		if idxOrg != idx {
			return false
		}

		sidebar := n.SideBar[idx]

		if sidebar.Disabled != s.Disabled.ValueBool() {
			return false
		}

		if sidebar.WidgetID != s.WidgetId.ValueString() {
			return false
		}

		if sidebar.WidgetNameSpace != s.WidgetNamespace.ValueString() {
			return false
		}

		a := make(map[string]string)

		s.Settings.Unmarshal(a)

		if !reflect.DeepEqual(sidebar.Settings, a) {
			return false
		}
	}

	return true
}

func (c *ContentType) DraftEditorInterface(n *contentful.EditorInterface) {
	n.Controls = pie.Map(c.Fields, func(field Field) contentful.Controls {

		control := contentful.Controls{
			FieldID: field.Id.ValueString(),
		}

		if field.Control != nil {
			control.WidgetID = field.Control.WidgetId.ValueStringPointer()
			control.WidgetNameSpace = field.Control.WidgetNamespace.ValueStringPointer()

			if field.Control.Settings != nil {
				control.Settings = field.Control.Settings.Draft()
			}
		}

		return control

	})

	n.SideBar = pie.Map(c.Sidebar, func(t Sidebar) contentful.Sidebar {

		sidebar := contentful.Sidebar{
			WidgetNameSpace: t.WidgetNamespace.ValueString(),
			WidgetID:        t.WidgetId.ValueString(),
			Disabled:        t.Disabled.ValueBool(),
		}

		if !sidebar.Disabled {
			settings := make(map[string]string)

			t.Settings.Unmarshal(settings)
			sidebar.Settings = settings
		}

		return sidebar
	})
}

func getValidations(contentfulValidations []model.FieldValidation) ([]Validation, error) {
	var validations []Validation

	for _, validation := range contentfulValidations {

		val, err := getValidation(validation)

		if err != nil {
			return nil, err
		}

		validations = append(validations, *val)
	}

	return validations, nil
}

func getValidation(cfVal model.FieldValidation) (*Validation, error) {

	if v, ok := cfVal.(model.FieldValidationPredefinedValues); ok {

		return &Validation{
			In: pie.Map(v.In, func(t any) types.String {
				return types.StringValue(t.(string))
			}),
		}, nil
	}

	if v, ok := cfVal.(model.FieldValidationUnique); ok {

		return &Validation{
			Unique: types.BoolValue(v.Unique),
		}, nil
	}

	if v, ok := cfVal.(model.FieldValidationRegex); ok {

		return &Validation{
			Regexp: &Regexp{
				Pattern: types.StringValue(v.Regex.Pattern),
			},
			Message: types.StringPointerValue(v.ErrorMessage),
		}, nil
	}

	if v, ok := cfVal.(model.FieldValidationSize); ok {

		return &Validation{
			Size: &Size{
				Max: types.Float64PointerValue(v.Size.Max),
				Min: types.Float64PointerValue(v.Size.Min),
			},
		}, nil
	}

	if v, ok := cfVal.(model.FieldValidationLink); ok {

		return &Validation{
			LinkContentType: pie.Map(v.LinkContentType, func(t string) types.String {
				return types.StringValue(t)
			}),
		}, nil
	}

	if v, ok := cfVal.(model.FieldValidationMimeType); ok {

		return &Validation{
			LinkMimetypeGroup: pie.Map(v.MimeTypes, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(v.ErrorMessage),
		}, nil
	}

	if v, ok := cfVal.(model.FieldValidationRange); ok {

		return &Validation{
			Range: &Size{
				Max: types.Float64PointerValue(v.Range.Max),
				Min: types.Float64PointerValue(v.Range.Min),
			},
		}, nil
	}

	if v, ok := cfVal.(model.FieldValidationEnabledNodeTypes); ok {

		return &Validation{
			EnabledNodeTypes: pie.Map(v.NodeTypes, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(v.ErrorMessage),
		}, nil
	}

	if v, ok := cfVal.(model.FieldValidationEnabledMarks); ok {

		return &Validation{
			EnabledMarks: pie.Map(v.Marks, func(t string) types.String {
				return types.StringValue(t)
			}),
			Message: types.StringPointerValue(v.ErrorMessage),
		}, nil
	}

	if v, ok := cfVal.(model.FieldValidationFileSize); ok {
		return &Validation{
			AssetFileSize: &Size{
				Max: types.Float64PointerValue(v.Size.Max),
				Min: types.Float64PointerValue(v.Size.Min),
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported validation used, %s. Please implement", reflect.TypeOf(cfVal).String())
}
