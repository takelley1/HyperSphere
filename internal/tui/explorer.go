// Path: internal/tui/explorer.go
// Description: Provide k9s-style command navigation, table interactions, and bulk actions.
package tui

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var (
	// ErrMissingCommandPrefix indicates a command missing ':' prefix.
	ErrMissingCommandPrefix = errors.New("command must start with ':'")
	// ErrUnknownResource indicates an unsupported resource view command.
	ErrUnknownResource = errors.New("unknown resource")
	// ErrUnsupportedHotKey indicates no table behavior exists for a key.
	ErrUnsupportedHotKey = errors.New("unsupported hotkey")
	// ErrInvalidAction indicates the action is unavailable for the active resource.
	ErrInvalidAction = errors.New("invalid action")
	// ErrNoPreviousView indicates no previous view exists for last-view toggling.
	ErrNoPreviousView = errors.New("no previous view")
	// ErrReadOnly indicates write actions are blocked by read-only mode.
	ErrReadOnly = errors.New("read-only mode")
	// ErrInvalidColumns indicates a requested column set is invalid.
	ErrInvalidColumns = errors.New("invalid columns")
)

// Resource identifies a table view namespace.
type Resource string

const (
	ResourceVM         Resource = "vm"
	ResourceLUN        Resource = "lun"
	ResourceCluster    Resource = "cluster"
	ResourceDatacenter Resource = "datacenter"
	ResourcePool       Resource = "resourcepool"
	ResourceNetwork    Resource = "network"
	ResourceTemplate   Resource = "template"
	ResourceSnapshot   Resource = "snapshot"
	ResourceTask       Resource = "task"
	ResourceEvent      Resource = "event"
	ResourceAlarm      Resource = "alarm"
	ResourceFolder     Resource = "folder"
	ResourceTag        Resource = "tag"
	ResourceHost       Resource = "host"
	ResourceDatastore  Resource = "datastore"
)

var resourceAliasMap = map[string]Resource{
	"vm":            ResourceVM,
	"vms":           ResourceVM,
	"lun":           ResourceLUN,
	"luns":          ResourceLUN,
	"cluster":       ResourceCluster,
	"clusters":      ResourceCluster,
	"cl":            ResourceCluster,
	"dc":            ResourceDatacenter,
	"datacenter":    ResourceDatacenter,
	"datacenters":   ResourceDatacenter,
	"rp":            ResourcePool,
	"resourcepool":  ResourcePool,
	"resourcepools": ResourcePool,
	"nw":            ResourceNetwork,
	"network":       ResourceNetwork,
	"networks":      ResourceNetwork,
	"tp":            ResourceTemplate,
	"template":      ResourceTemplate,
	"templates":     ResourceTemplate,
	"ss":            ResourceSnapshot,
	"snap":          ResourceSnapshot,
	"snapshot":      ResourceSnapshot,
	"snapshots":     ResourceSnapshot,
	"task":          ResourceTask,
	"tasks":         ResourceTask,
	"event":         ResourceEvent,
	"events":        ResourceEvent,
	"alarm":         ResourceAlarm,
	"alarms":        ResourceAlarm,
	"folder":        ResourceFolder,
	"folders":       ResourceFolder,
	"tag":           ResourceTag,
	"tags":          ResourceTag,
	"host":          ResourceHost,
	"hosts":         ResourceHost,
	"datastore":     ResourceDatastore,
	"datastores":    ResourceDatastore,
	"ds":            ResourceDatastore,
}

// VMRow represents one VM row in the resource table.
type VMRow struct {
	Name            string
	Tags            string
	Cluster         string
	Host            string
	Network         string
	PowerState      string
	Datastore       string
	AttachedStorage string
	IPAddress       string
	DNSName         string
	CPUCount        int
	MemoryMB        int
	UsedCPUPercent  int
	UsedMemoryMB    int
	UsedStorageGB   int
	LargestDiskGB   int
	SnapshotTotalGB int
	Owner           string
	Comments        string
	Description     string
	SnapshotCount   int
	Snapshots       []VMSnapshot
}

// VMSnapshot stores summary fields for one VM snapshot.
type VMSnapshot struct {
	Identifier string
	Timestamp  string
}

// LUNRow represents one LUN row in the resource table.
type LUNRow struct {
	Name       string
	Tags       string
	Cluster    string
	Datastore  string
	CapacityGB int
	UsedGB     int
}

// ClusterRow represents one cluster row in the resource table.
type ClusterRow struct {
	Name              string
	Tags              string
	Datacenter        string
	Hosts             int
	VMCount           int
	CPUUsagePercent   int
	MemUsagePercent   int
	ResourcePoolCount int
	NetworkCount      int
}

// DatacenterRow represents one datacenter row in the resource table.
type DatacenterRow struct {
	Name            string
	ClusterCount    int
	HostCount       int
	VMCount         int
	DatastoreCount  int
	CPUUsagePercent int
	MemUsagePercent int
}

// ResourcePoolRow represents one resource pool row in the resource table.
type ResourcePoolRow struct {
	Name              string
	Cluster           string
	CPUReservationMHz int
	MemReservationMB  int
	VMCount           int
	CPULimitMHz       int
	MemLimitMB        int
}

// NetworkRow represents one network row in the resource table.
type NetworkRow struct {
	Name        string
	Type        string
	VLAN        string
	Switch      string
	AttachedVMs int
	MTU         int
	Uplinks     int
}

// TemplateRow represents one VM template row in the resource table.
type TemplateRow struct {
	Name      string
	OS        string
	Datastore string
	Folder    string
	Age       string
	CPUCount  int
	MemoryMB  int
}

// SnapshotRow represents one VM snapshot row in the resource table.
type SnapshotRow struct {
	VM       string
	Snapshot string
	Size     string
	Created  string
	Age      string
	Quiesced string
	Owner    string
}

// TaskRow represents one vCenter task stream row in the resource table.
type TaskRow struct {
	Entity   string
	Action   string
	State    string
	Started  string
	Duration string
	Owner    string
}

// EventRow represents one inventory event stream row in the resource table.
type EventRow struct {
	Time     string
	Severity string
	Entity   string
	Message  string
	User     string
}

// AlarmRow represents one active alarm row in the resource table.
type AlarmRow struct {
	Entity    string
	Alarm     string
	Status    string
	Triggered string
	AckedBy   string
}

// FolderRow represents one inventory folder row in the resource table.
type FolderRow struct {
	Path     string
	Type     string
	Children int
	VMCount  int
}

// TagRow represents one tag/category row in the resource table.
type TagRow struct {
	Tag             string
	Category        string
	Cardinality     string
	AttachedObjects int
}

// HostRow represents one host row in the resource table.
type HostRow struct {
	Name            string
	Tags            string
	Cluster         string
	CPUUsagePercent int
	MemUsagePercent int
	ConnectionState string
	CoreCount       int
	ThreadCount     int
	VMCount         int
}

// DatastoreRow represents one datastore row in the resource table.
type DatastoreRow struct {
	Name       string
	Tags       string
	Cluster    string
	CapacityGB int
	UsedGB     int
	FreeGB     int
	Type       string
	LatencyMS  int
}

// Catalog stores rows available for each resource view.
type Catalog struct {
	VMs           []VMRow
	LUNs          []LUNRow
	Clusters      []ClusterRow
	Datacenters   []DatacenterRow
	ResourcePools []ResourcePoolRow
	Networks      []NetworkRow
	Templates     []TemplateRow
	Snapshots     []SnapshotRow
	Tasks         []TaskRow
	Events        []EventRow
	Alarms        []AlarmRow
	Folders       []FolderRow
	Tags          []TagRow
	Hosts         []HostRow
	Datastores    []DatastoreRow
}

// ResourceView stores a concrete table to render.
type ResourceView struct {
	Resource    Resource
	Columns     []string
	Rows        [][]string
	IDs         []string
	SortHotKeys map[string]string
	Actions     []string
}

// DetailField stores one key-value pair for a selected resource detail view.
type DetailField struct {
	Key   string
	Value string
}

// ResourceDetails stores all fields rendered by the describe panel.
type ResourceDetails struct {
	Title  string
	Fields []DetailField
}

// ActionExecutor applies bulk actions through a VMware API adapter.
type ActionExecutor interface {
	Execute(resource Resource, action string, ids []string) error
}

// Navigator handles command-mode resource switching.
type Navigator struct {
	catalog Catalog
	active  Resource
}

// Session tracks interactive table state for one active view.
type Session struct {
	navigator       Navigator
	view            ResourceView
	baseView        ResourceView
	previousView    Resource
	selectedRow     int
	selectedColumn  int
	sortColumn      string
	sortAsc         bool
	filterText      string
	readOnly        bool
	marks           map[string]struct{}
	markAnchor      int
	columnSelection map[Resource][]string
}

// NewNavigator builds a command navigator with a VM default view.
func NewNavigator(catalog Catalog) Navigator {
	return Navigator{catalog: catalog, active: ResourceVM}
}

// NewSession initializes an interactive table session.
func NewSession(catalog Catalog) Session {
	navigator := NewNavigator(catalog)
	view, _ := navigator.TableFor(ResourceVM)
	return Session{
		navigator:       navigator,
		view:            view,
		baseView:        view,
		marks:           map[string]struct{}{},
		markAnchor:      -1,
		columnSelection: map[Resource][]string{},
	}
}

// ActiveView returns the currently selected resource view.
func (n *Navigator) ActiveView() Resource {
	return n.active
}

// Execute parses a command and switches the active resource view.
func (n *Navigator) Execute(command string) (ResourceView, error) {
	resource, err := parseCommand(command)
	if err != nil {
		return ResourceView{}, err
	}
	view, _ := n.viewFor(resource)
	n.active = resource
	return view, nil
}

// TableFor builds a table view for a specific resource.
func (n *Navigator) TableFor(resource Resource) (ResourceView, error) {
	view, ok := n.viewFor(resource)
	if ok {
		return view, nil
	}
	return ResourceView{}, fmt.Errorf("%w: %s", ErrUnknownResource, resource)
}

func (n *Navigator) viewFor(resource Resource) (ResourceView, bool) {
	switch resource {
	case ResourceVM:
		return vmView(n.catalog.VMs), true
	case ResourceLUN:
		return lunView(n.catalog.LUNs), true
	case ResourceCluster:
		return clusterView(n.catalog.Clusters), true
	case ResourceDatacenter:
		return datacenterView(n.catalog.Datacenters), true
	case ResourcePool:
		return resourcePoolView(n.catalog.ResourcePools), true
	case ResourceNetwork:
		return networkView(n.catalog.Networks), true
	case ResourceTemplate:
		return templateView(n.catalog.Templates), true
	case ResourceSnapshot:
		return snapshotView(n.catalog.Snapshots), true
	case ResourceTask:
		return taskView(n.catalog.Tasks), true
	case ResourceEvent:
		return eventView(n.catalog.Events), true
	case ResourceAlarm:
		return alarmView(n.catalog.Alarms), true
	case ResourceFolder:
		return folderView(n.catalog.Folders), true
	case ResourceTag:
		return tagView(n.catalog.Tags), true
	case ResourceHost:
		return hostView(n.catalog.Hosts), true
	case ResourceDatastore:
		return datastoreView(n.catalog.Datastores), true
	default:
		return ResourceView{}, false
	}
}

// ExecuteCommand switches the active session view from a ':' command.
func (s *Session) ExecuteCommand(command string) error {
	view, err := s.navigator.Execute(command)
	if err != nil {
		return err
	}
	view, err = s.applyStoredColumns(view)
	if err != nil {
		return err
	}
	if s.view.Resource != view.Resource {
		s.previousView = s.view.Resource
	}
	s.view = view
	s.baseView = view
	s.selectedRow = 0
	s.selectedColumn = 0
	s.sortColumn = ""
	s.sortAsc = true
	s.filterText = ""
	s.marks = map[string]struct{}{}
	s.markAnchor = -1
	return nil
}

// SetVisibleColumns sets the visible columns for the current view and persists the selection per view.
func (s *Session) SetVisibleColumns(columns []string) error {
	resource := s.view.Resource
	fullView, err := s.navigator.TableFor(resource)
	if err != nil {
		return err
	}
	filteredView, normalized, err := selectVisibleColumns(fullView, columns)
	if err != nil {
		return err
	}
	s.columnSelection[resource] = normalized
	s.view = filteredView
	s.baseView = filteredView
	s.sortColumn = ""
	s.sortAsc = true
	s.selectedColumn = clampSelectionIndex(s.selectedColumn, len(s.view.Columns))
	if s.filterText != "" {
		s.ApplyFilter(s.filterText)
	}
	return nil
}

// ResetVisibleColumns clears custom visible columns for the current view.
func (s *Session) ResetVisibleColumns() error {
	resource := s.view.Resource
	delete(s.columnSelection, resource)
	fullView, err := s.navigator.TableFor(resource)
	if err != nil {
		return err
	}
	s.view = fullView
	s.baseView = fullView
	s.sortColumn = ""
	s.sortAsc = true
	s.selectedColumn = clampSelectionIndex(s.selectedColumn, len(s.view.Columns))
	if s.filterText != "" {
		s.ApplyFilter(s.filterText)
	}
	return nil
}

// VisibleColumns returns the currently displayed columns.
func (s *Session) VisibleColumns() []string {
	return append([]string{}, s.view.Columns...)
}

// AvailableColumns returns the full column set for the active resource.
func (s *Session) AvailableColumns() ([]string, error) {
	fullView, err := s.navigator.TableFor(s.view.Resource)
	if err != nil {
		return nil, err
	}
	return append([]string{}, fullView.Columns...), nil
}

// HandleKey applies one k9s-inspired table hotkey.
func (s *Session) HandleKey(key string) error {
	normalized := normalizeKey(key)
	if normalized == "" {
		return nil
	}
	if tryMoveRow(s, normalized) || tryMoveColumn(s, normalized) {
		return nil
	}
	if normalized == "SPACE" {
		s.toggleMark()
		return nil
	}
	if normalized == "CTRL+SPACE" {
		s.spanMark()
		return nil
	}
	if normalized == "CTRL+\\" {
		s.clearMarks()
		return nil
	}
	if normalized == "O" || normalized == "SHIFT+O" {
		return s.sortBySelectedColumn()
	}
	if normalized == "SHIFT+I" {
		return s.invertSort()
	}
	column, ok := s.view.SortHotKeys[normalized]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnsupportedHotKey, key)
	}
	s.sortByColumn(column, true)
	return nil
}

// ApplyAction executes a resource-specific action across selected rows.
func (s *Session) ApplyAction(action string, executor ActionExecutor) error {
	if s.readOnly {
		return ErrReadOnly
	}
	normalized := strings.ToLower(strings.TrimSpace(action))
	if !containsAction(s.view.Actions, normalized) {
		return fmt.Errorf("%w: %s", ErrInvalidAction, action)
	}
	ids := s.selectedIDs()
	if len(ids) == 0 {
		return fmt.Errorf("%w: no selected rows", ErrInvalidAction)
	}
	return executor.Execute(s.view.Resource, normalized, ids)
}

// CurrentView returns the current table snapshot.
func (s *Session) CurrentView() ResourceView {
	return s.view
}

// SelectedResourceDetails builds describe-panel fields for the selected row.
func (s *Session) SelectedResourceDetails() (ResourceDetails, error) {
	id, rowIndex, err := s.selectedRowContext()
	if err != nil {
		return ResourceDetails{}, err
	}
	if s.view.Resource == ResourceVM {
		return s.vmDetailsByID(id)
	}
	return genericDetailsFromRow(s.view, rowIndex), nil
}

// SelectedRow returns the currently focused row index.
func (s *Session) SelectedRow() int {
	return s.selectedRow
}

// SelectedColumn returns the currently focused column index.
func (s *Session) SelectedColumn() int {
	return s.selectedColumn
}

// SetSelection updates focused row and column with bounds clamping.
func (s *Session) SetSelection(row int, column int) {
	s.selectedRow = clampSelectionIndex(row, len(s.view.Rows))
	s.selectedColumn = clampSelectionIndex(column, len(s.view.Columns))
}

// IsMarked reports whether a specific row ID is marked.
func (s *Session) IsMarked(id string) bool {
	_, ok := s.marks[id]
	return ok
}

// ReadOnly returns whether mutating actions are blocked.
func (s *Session) ReadOnly() bool {
	return s.readOnly
}

// SetReadOnly toggles read-only mode for mutating actions.
func (s *Session) SetReadOnly(readOnly bool) {
	s.readOnly = readOnly
}

// ApplyFilter filters rows by substring match across all columns.
func (s *Session) ApplyFilter(filter string) {
	s.filterText = strings.ToLower(strings.TrimSpace(filter))
	if s.filterText == "" {
		s.view = s.baseView
		s.clampSelectedRow()
		return
	}
	s.view = filterView(s.baseView, s.filterText)
	s.clampSelectedRow()
}

// LastView toggles back to the previous resource view.
func (s *Session) LastView() error {
	if s.previousView == "" {
		return ErrNoPreviousView
	}
	current := s.view.Resource
	if err := s.ExecuteCommand(":" + string(s.previousView)); err != nil {
		return err
	}
	s.previousView = current
	return nil
}

// Render returns a k9s-style interactive table representation.
func (s *Session) Render() string {
	return RenderInteractiveView(
		s.view,
		s.selectedRow,
		s.selectedColumn,
		s.sortColumn,
		s.sortAsc,
		s.marks,
		s.readOnly,
	)
}

// RenderResourceView renders a static table view.
func RenderResourceView(view ResourceView) string {
	return RenderInteractiveView(view, 0, 0, "", true, map[string]struct{}{}, false)
}

// RenderInteractiveView renders a table with sort and mark indicators.
func RenderInteractiveView(
	view ResourceView,
	selectedRow int,
	selectedColumn int,
	sortColumn string,
	sortAsc bool,
	marks map[string]struct{},
	readOnly bool,
) string {
	builder := &strings.Builder{}
	builder.WriteString(headerLine(view.Resource, sortColumn, sortAsc, len(marks), readOnly))
	if len(view.Rows) == 0 {
		builder.WriteString("No resources found.\n")
		return builder.String()
	}
	columns := append([]string{"M", ">"}, view.Columns...)
	rows := decorateRows(view, selectedRow, marks)
	rows = viewportRows(rows, selectedRow)
	widths := columnWidths(columns, rows)
	builder.WriteString(
		formatCells(markColumn(columns, selectedColumn+2, sortColumn, sortAsc), widths),
	)
	for _, row := range rows {
		builder.WriteString(formatCells(row, widths))
	}
	builder.WriteString(actionLine(view.Actions))
	builder.WriteString("Keys: Space mark | J/K or Up/Down row | Shift+Left/Right column | Shift+O sort column\n")
	return builder.String()
}

func headerLine(
	resource Resource,
	sortColumn string,
	sortAsc bool,
	marked int,
	readOnly bool,
) string {
	mode := "RW"
	if readOnly {
		mode = "RO"
	}
	if sortColumn == "" {
		return fmt.Sprintf(
			"HyperSphere :: %s | Mode: %s | Sort: - | Marks[%d]\n",
			resource,
			mode,
			marked,
		)
	}
	arrow := "↑"
	if !sortAsc {
		arrow = "↓"
	}
	return fmt.Sprintf(
		"HyperSphere :: %s | Mode: %s | Sort: %s%s | Marks[%d]\n",
		resource,
		mode,
		sortColumn,
		arrow,
		marked,
	)
}

func actionLine(actions []string) string {
	if len(actions) == 0 {
		return "Actions: none\n"
	}
	return "Actions: " + strings.Join(actions, ", ") + " (run: !<action>)\n"
}

func markColumn(
	columns []string,
	selectedColumn int,
	sortColumn string,
	sortAsc bool,
) []string {
	decorated := make([]string, len(columns))
	copy(decorated, columns)
	for index, column := range columns {
		label := decorated[index]
		if sortColumn != "" && column == sortColumn {
			label += sortDirectionGlyph(sortAsc)
		}
		if index == selectedColumn {
			label = "[" + label + "]"
		}
		decorated[index] = label
	}
	return decorated
}

func sortDirectionGlyph(sortAsc bool) string {
	if sortAsc {
		return "↑"
	}
	return "↓"
}

func decorateRows(view ResourceView, selectedRow int, marks map[string]struct{}) [][]string {
	rows := make([][]string, 0, len(view.Rows))
	for index, row := range view.Rows {
		rows = append(rows, decorateRow(view, row, index, selectedRow, marks))
	}
	return rows
}

func decorateRow(
	view ResourceView,
	row []string,
	index int,
	selectedRow int,
	marks map[string]struct{},
) []string {
	mark := " "
	if index < len(view.IDs) {
		if _, ok := marks[view.IDs[index]]; ok {
			mark = "*"
		}
	}
	cursor := " "
	if index == selectedRow {
		cursor = ">"
	}
	cells := make([]string, 0, len(row)+2)
	cells = append(cells, mark, cursor)
	cells = append(cells, row...)
	return cells
}

func parseCommand(command string) (Resource, error) {
	trimmed := strings.TrimSpace(command)
	if !strings.HasPrefix(trimmed, ":") {
		return "", ErrMissingCommandPrefix
	}
	name := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, ":")))
	name = firstField(name)
	if name == "" {
		return "", fmt.Errorf("%w: empty", ErrUnknownResource)
	}
	resource, ok := normalizeResourceName(name)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnknownResource, name)
	}
	return resource, nil
}

func firstField(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func vmView(rows []VMRow) ResourceView {
	columns := []string{
		"NAME",
		"POWER",
		"USED_CPU_PERCENT",
		"USED_MEMORY_MB",
		"USED_STORAGE_GB",
		"IP_ADDRESS",
		"DNS_NAME",
		"CLUSTER",
		"HOST",
		"NETWORK",
		"TOTAL_CPU_CORES",
		"TOTAL_RAM_MB",
		"LARGEST_DISK_GB",
		"SNAPSHOT_COUNT",
		"SNAPSHOT_TOTAL_GB",
		"ATTACHED_STORAGE",
	}
	return buildView(ResourceVM, columns, vmSortHotKeys(), vmActions(), rows, vmCells)
}

func lunView(rows []LUNRow) ResourceView {
	columns := []string{
		"NAME",
		"TAGS",
		"CLUSTER",
		"DATASTORE",
		"CAPACITY_GB",
		"USED_GB",
		"FREE_GB",
		"UTIL_PERCENT",
	}
	return buildView(ResourceLUN, columns, lunSortHotKeys(), lunActions(), rows, lunCells)
}

func clusterView(rows []ClusterRow) ResourceView {
	columns := []string{
		"NAME",
		"TAGS",
		"DATACENTER",
		"HOSTS",
		"VMS",
		"CPU_PERCENT",
		"MEM_PERCENT",
		"RESOURCE_POOLS",
		"NETWORKS",
	}
	return buildView(ResourceCluster, columns, clusterSortHotKeys(), clusterActions(), rows, clusterCells)
}

func datacenterView(rows []DatacenterRow) ResourceView {
	columns := []string{"NAME", "CLUSTERS", "HOSTS", "VMS", "DATASTORES", "CPU_PERCENT", "MEM_PERCENT"}
	return buildView(
		ResourceDatacenter,
		columns,
		datacenterSortHotKeys(),
		datacenterActions(),
		rows,
		datacenterCells,
	)
}

func resourcePoolView(rows []ResourcePoolRow) ResourceView {
	columns := []string{"NAME", "CLUSTER", "CPU_RES", "MEM_RES", "VM_COUNT", "CPU_LIMIT", "MEM_LIMIT"}
	return buildView(
		ResourcePool,
		columns,
		resourcePoolSortHotKeys(),
		resourcePoolActions(),
		rows,
		resourcePoolCells,
	)
}

func networkView(rows []NetworkRow) ResourceView {
	columns := []string{"NAME", "TYPE", "VLAN", "SWITCH", "ATTACHED_VMS", "MTU", "UPLINKS"}
	return buildView(
		ResourceNetwork,
		columns,
		networkSortHotKeys(),
		networkActions(),
		rows,
		networkCells,
	)
}

func templateView(rows []TemplateRow) ResourceView {
	columns := []string{"NAME", "OS", "DATASTORE", "FOLDER", "AGE", "CPU_COUNT", "MEMORY_MB"}
	return buildView(
		ResourceTemplate,
		columns,
		templateSortHotKeys(),
		templateActions(),
		rows,
		templateCells,
	)
}

func snapshotView(rows []SnapshotRow) ResourceView {
	columns := []string{"VM", "SNAPSHOT", "SIZE", "CREATED", "AGE", "QUIESCED", "OWNER"}
	return buildView(
		ResourceSnapshot,
		columns,
		snapshotSortHotKeys(),
		snapshotActions(),
		rows,
		snapshotCells,
	)
}

func taskView(rows []TaskRow) ResourceView {
	columns := []string{"ENTITY", "ACTION", "STATE", "STARTED", "DURATION", "OWNER"}
	return buildView(ResourceTask, columns, taskSortHotKeys(), taskActions(), rows, taskCells)
}

func eventView(rows []EventRow) ResourceView {
	columns := []string{"TIME", "SEVERITY", "ENTITY", "MESSAGE", "USER"}
	return buildView(ResourceEvent, columns, eventSortHotKeys(), eventActions(), rows, eventCells)
}

func alarmView(rows []AlarmRow) ResourceView {
	columns := []string{"ENTITY", "ALARM", "STATUS", "TRIGGERED", "ACKED_BY"}
	return buildView(ResourceAlarm, columns, alarmSortHotKeys(), alarmActions(), rows, alarmCells)
}

func folderView(rows []FolderRow) ResourceView {
	columns := []string{"PATH", "TYPE", "CHILDREN", "VM_COUNT"}
	return buildView(ResourceFolder, columns, folderSortHotKeys(), folderActions(), rows, folderCells)
}

func tagView(rows []TagRow) ResourceView {
	columns := []string{"TAG", "CATEGORY", "CARDINALITY", "ATTACHED_OBJECTS"}
	return buildView(ResourceTag, columns, tagSortHotKeys(), tagActions(), rows, tagCells)
}

func hostView(rows []HostRow) ResourceView {
	columns := []string{
		"NAME",
		"TAGS",
		"CLUSTER",
		"CPU_PERCENT",
		"MEM_PERCENT",
		"CONNECTION",
		"CORES",
		"THREADS",
		"VMS",
	}
	return buildView(ResourceHost, columns, hostSortHotKeys(), hostActions(), rows, hostCells)
}

func datastoreView(rows []DatastoreRow) ResourceView {
	columns := []string{
		"NAME",
		"TAGS",
		"CLUSTER",
		"CAPACITY_GB",
		"USED_GB",
		"FREE_GB",
		"TYPE",
		"LATENCY_MS",
	}
	return buildView(
		ResourceDatastore,
		columns,
		datastoreSortHotKeys(),
		datastoreActions(),
		rows,
		datastoreCells,
	)
}

func buildView[T any](
	resource Resource,
	columns []string,
	sortHotKeys map[string]string,
	actions []string,
	rows []T,
	toCells func(T) (string, []string),
) ResourceView {
	ids := make([]string, 0, len(rows))
	viewRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		id, cells := toCells(row)
		ids = append(ids, id)
		viewRows = append(viewRows, cells)
	}
	return ResourceView{Resource: resource, Columns: columns, Rows: viewRows, IDs: ids, SortHotKeys: sortHotKeys, Actions: actions}
}

func vmCells(row VMRow) (string, []string) {
	attachedStorage := defaultCell(row.AttachedStorage)
	if attachedStorage == "-" {
		attachedStorage = defaultCell(row.Datastore)
	}
	return row.Name, []string{
		row.Name,
		defaultCell(row.PowerState),
		strconv.Itoa(row.UsedCPUPercent),
		strconv.Itoa(row.UsedMemoryMB),
		strconv.Itoa(row.UsedStorageGB),
		defaultCell(row.IPAddress),
		defaultCell(row.DNSName),
		defaultCell(row.Cluster),
		defaultCell(row.Host),
		defaultCell(row.Network),
		strconv.Itoa(row.CPUCount),
		strconv.Itoa(row.MemoryMB),
		strconv.Itoa(row.LargestDiskGB),
		strconv.Itoa(vmSnapshotCount(row)),
		strconv.Itoa(row.SnapshotTotalGB),
		attachedStorage,
	}
}

func lunCells(row LUNRow) (string, []string) {
	freeGB := row.CapacityGB - row.UsedGB
	if freeGB < 0 {
		freeGB = 0
	}
	return row.Name, []string{
		row.Name,
		defaultCell(row.Tags),
		defaultCell(row.Cluster),
		defaultCell(row.Datastore),
		strconv.Itoa(row.CapacityGB),
		strconv.Itoa(row.UsedGB),
		strconv.Itoa(freeGB),
		strconv.Itoa(lunUtilPercent(row.CapacityGB, row.UsedGB)),
	}
}

func clusterCells(row ClusterRow) (string, []string) {
	return row.Name, []string{
		row.Name,
		defaultCell(row.Tags),
		defaultCell(row.Datacenter),
		strconv.Itoa(row.Hosts),
		strconv.Itoa(row.VMCount),
		strconv.Itoa(row.CPUUsagePercent),
		strconv.Itoa(row.MemUsagePercent),
		strconv.Itoa(row.ResourcePoolCount),
		strconv.Itoa(row.NetworkCount),
	}
}

func hostCells(row HostRow) (string, []string) {
	return row.Name, []string{
		row.Name,
		defaultCell(row.Tags),
		defaultCell(row.Cluster),
		strconv.Itoa(row.CPUUsagePercent),
		strconv.Itoa(row.MemUsagePercent),
		defaultCell(row.ConnectionState),
		strconv.Itoa(row.CoreCount),
		strconv.Itoa(row.ThreadCount),
		strconv.Itoa(row.VMCount),
	}
}

func datacenterCells(row DatacenterRow) (string, []string) {
	return row.Name, []string{
		row.Name,
		strconv.Itoa(row.ClusterCount),
		strconv.Itoa(row.HostCount),
		strconv.Itoa(row.VMCount),
		strconv.Itoa(row.DatastoreCount),
		strconv.Itoa(row.CPUUsagePercent),
		strconv.Itoa(row.MemUsagePercent),
	}
}

func resourcePoolCells(row ResourcePoolRow) (string, []string) {
	return row.Name, []string{
		row.Name,
		defaultCell(row.Cluster),
		strconv.Itoa(row.CPUReservationMHz),
		strconv.Itoa(row.MemReservationMB),
		strconv.Itoa(row.VMCount),
		strconv.Itoa(row.CPULimitMHz),
		strconv.Itoa(row.MemLimitMB),
	}
}

func networkCells(row NetworkRow) (string, []string) {
	return row.Name, []string{
		row.Name,
		defaultCell(row.Type),
		defaultCell(row.VLAN),
		defaultCell(row.Switch),
		strconv.Itoa(row.AttachedVMs),
		strconv.Itoa(row.MTU),
		strconv.Itoa(row.Uplinks),
	}
}

func templateCells(row TemplateRow) (string, []string) {
	return row.Name, []string{
		row.Name,
		defaultCell(row.OS),
		defaultCell(row.Datastore),
		defaultCell(row.Folder),
		defaultCell(row.Age),
		strconv.Itoa(row.CPUCount),
		strconv.Itoa(row.MemoryMB),
	}
}

func snapshotCells(row SnapshotRow) (string, []string) {
	id := defaultCell(row.VM) + ":" + defaultCell(row.Snapshot)
	return id, []string{
		defaultCell(row.VM),
		defaultCell(row.Snapshot),
		defaultCell(row.Size),
		defaultCell(row.Created),
		defaultCell(row.Age),
		defaultCell(row.Quiesced),
		defaultCell(row.Owner),
	}
}

func taskCells(row TaskRow) (string, []string) {
	id := defaultCell(row.Entity) + ":" + defaultCell(row.Action) + ":" + defaultCell(row.Started)
	return id, []string{
		defaultCell(row.Entity),
		defaultCell(row.Action),
		defaultCell(row.State),
		defaultCell(row.Started),
		defaultCell(row.Duration),
		defaultCell(row.Owner),
	}
}

func eventCells(row EventRow) (string, []string) {
	id := defaultCell(row.Time) + ":" + defaultCell(row.Entity) + ":" + defaultCell(row.Message)
	return id, []string{
		defaultCell(row.Time),
		defaultCell(row.Severity),
		defaultCell(row.Entity),
		defaultCell(row.Message),
		defaultCell(row.User),
	}
}

func alarmCells(row AlarmRow) (string, []string) {
	id := defaultCell(row.Entity) + ":" + defaultCell(row.Alarm)
	return id, []string{
		defaultCell(row.Entity),
		defaultCell(row.Alarm),
		defaultCell(row.Status),
		defaultCell(row.Triggered),
		defaultCell(row.AckedBy),
	}
}

func folderCells(row FolderRow) (string, []string) {
	return defaultCell(row.Path), []string{
		defaultCell(row.Path),
		defaultCell(row.Type),
		strconv.Itoa(row.Children),
		strconv.Itoa(row.VMCount),
	}
}

func tagCells(row TagRow) (string, []string) {
	return defaultCell(row.Tag), []string{
		defaultCell(row.Tag),
		defaultCell(row.Category),
		defaultCell(row.Cardinality),
		strconv.Itoa(row.AttachedObjects),
	}
}

func datastoreCells(row DatastoreRow) (string, []string) {
	return row.Name, []string{
		row.Name,
		defaultCell(row.Tags),
		defaultCell(row.Cluster),
		strconv.Itoa(row.CapacityGB),
		strconv.Itoa(row.UsedGB),
		strconv.Itoa(row.FreeGB),
		defaultCell(row.Type),
		strconv.Itoa(row.LatencyMS),
	}
}

func defaultCell(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func lunUtilPercent(capacityGB int, usedGB int) int {
	if capacityGB <= 0 {
		return 0
	}
	percent := (usedGB * 100) / capacityGB
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func vmSortHotKeys() map[string]string {
	return map[string]string{
		"N": "NAME",
		"P": "POWER",
		"U": "USED_CPU_PERCENT",
		"M": "USED_MEMORY_MB",
		"G": "USED_STORAGE_GB",
		"I": "IP_ADDRESS",
		"D": "DNS_NAME",
		"C": "CLUSTER",
		"H": "HOST",
		"W": "NETWORK",
		"T": "TOTAL_CPU_CORES",
		"R": "TOTAL_RAM_MB",
		"L": "LARGEST_DISK_GB",
		"S": "SNAPSHOT_COUNT",
		"Z": "SNAPSHOT_TOTAL_GB",
		"A": "ATTACHED_STORAGE",
	}
}

func lunSortHotKeys() map[string]string {
	return map[string]string{
		"N": "NAME",
		"T": "TAGS",
		"C": "CLUSTER",
		"D": "DATASTORE",
		"G": "CAPACITY_GB",
		"U": "USED_GB",
		"F": "FREE_GB",
		"P": "UTIL_PERCENT",
	}
}

func clusterSortHotKeys() map[string]string {
	return map[string]string{
		"N": "NAME",
		"T": "TAGS",
		"D": "DATACENTER",
		"H": "HOSTS",
		"V": "VMS",
		"C": "CPU_PERCENT",
		"M": "MEM_PERCENT",
		"R": "RESOURCE_POOLS",
		"W": "NETWORKS",
	}
}

func datacenterSortHotKeys() map[string]string {
	return map[string]string{
		"N": "NAME",
		"C": "CLUSTERS",
		"H": "HOSTS",
		"V": "VMS",
		"D": "DATASTORES",
		"P": "CPU_PERCENT",
		"M": "MEM_PERCENT",
	}
}

func resourcePoolSortHotKeys() map[string]string {
	return map[string]string{
		"N": "NAME",
		"C": "CLUSTER",
		"P": "CPU_RES",
		"M": "MEM_RES",
		"V": "VM_COUNT",
		"L": "CPU_LIMIT",
		"R": "MEM_LIMIT",
	}
}

func networkSortHotKeys() map[string]string {
	return map[string]string{
		"N": "NAME",
		"T": "TYPE",
		"V": "VLAN",
		"S": "SWITCH",
		"A": "ATTACHED_VMS",
		"M": "MTU",
		"U": "UPLINKS",
	}
}

func templateSortHotKeys() map[string]string {
	return map[string]string{
		"N": "NAME",
		"O": "OS",
		"D": "DATASTORE",
		"F": "FOLDER",
		"A": "AGE",
		"C": "CPU_COUNT",
		"M": "MEMORY_MB",
	}
}

func snapshotSortHotKeys() map[string]string {
	return map[string]string{
		"V": "VM",
		"S": "SNAPSHOT",
		"Z": "SIZE",
		"C": "CREATED",
		"A": "AGE",
		"Q": "QUIESCED",
		"O": "OWNER",
	}
}

func taskSortHotKeys() map[string]string {
	return map[string]string{
		"E": "ENTITY",
		"A": "ACTION",
		"S": "STATE",
		"T": "STARTED",
		"D": "DURATION",
		"O": "OWNER",
	}
}

func eventSortHotKeys() map[string]string {
	return map[string]string{
		"T": "TIME",
		"S": "SEVERITY",
		"E": "ENTITY",
		"M": "MESSAGE",
		"U": "USER",
	}
}

func alarmSortHotKeys() map[string]string {
	return map[string]string{
		"E": "ENTITY",
		"A": "ALARM",
		"S": "STATUS",
		"T": "TRIGGERED",
		"K": "ACKED_BY",
	}
}

func folderSortHotKeys() map[string]string {
	return map[string]string{
		"P": "PATH",
		"T": "TYPE",
		"C": "CHILDREN",
		"V": "VM_COUNT",
	}
}

func tagSortHotKeys() map[string]string {
	return map[string]string{
		"T": "TAG",
		"C": "CATEGORY",
		"R": "CARDINALITY",
		"A": "ATTACHED_OBJECTS",
	}
}

func hostSortHotKeys() map[string]string {
	return map[string]string{
		"N": "NAME",
		"T": "TAGS",
		"C": "CLUSTER",
		"P": "CPU_PERCENT",
		"M": "MEM_PERCENT",
		"S": "CONNECTION",
		"O": "CORES",
		"H": "THREADS",
		"V": "VMS",
	}
}

func datastoreSortHotKeys() map[string]string {
	return map[string]string{
		"N": "NAME",
		"T": "TAGS",
		"C": "CLUSTER",
		"A": "CAPACITY_GB",
		"U": "USED_GB",
		"F": "FREE_GB",
		"Y": "TYPE",
		"L": "LATENCY_MS",
	}
}

func vmActions() []string {
	return []string{"power-on", "power-off", "migrate", "edit-tags"}
}

func lunActions() []string {
	return []string{"rescan", "expand", "edit-tags"}
}

func clusterActions() []string {
	return []string{"enter-maintenance", "exit-maintenance", "rebalance", "edit-tags"}
}

func hostActions() []string {
	return []string{"enter-maintenance", "exit-maintenance", "disconnect", "reconnect", "edit-tags"}
}

func datacenterActions() []string {
	return []string{"refresh", "edit-tags"}
}

func resourcePoolActions() []string {
	return []string{"set-reservation", "rebalance", "edit-tags"}
}

func networkActions() []string {
	return []string{"attach-vm", "detach-vm", "edit-tags"}
}

func templateActions() []string {
	return []string{"clone", "edit-tags"}
}

func snapshotActions() []string {
	return []string{"revert", "delete", "edit-tags"}
}

func taskActions() []string {
	return []string{"cancel", "retry"}
}

func eventActions() []string {
	return []string{"acknowledge"}
}

func alarmActions() []string {
	return []string{"acknowledge", "clear"}
}

func folderActions() []string {
	return []string{"open", "rename"}
}

func tagActions() []string {
	return []string{"assign", "unassign"}
}

func datastoreActions() []string {
	return []string{"rescan", "rebalance", "mount", "unmount", "edit-tags"}
}

func normalizeKey(key string) string {
	if key == " " {
		return "SPACE"
	}
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	return strings.ToUpper(trimmed)
}

func tryMoveRow(session *Session, key string) bool {
	if key == "DOWN" || key == "J" {
		session.moveRow(1)
		return true
	}
	if key == "UP" || key == "K" {
		session.moveRow(-1)
		return true
	}
	return false
}

func tryMoveColumn(session *Session, key string) bool {
	if key == "SHIFT+RIGHT" || key == "RIGHT" {
		session.moveColumn(1)
		return true
	}
	if key == "SHIFT+LEFT" || key == "LEFT" {
		session.moveColumn(-1)
		return true
	}
	return false
}

func (s *Session) moveRow(delta int) {
	if len(s.view.Rows) == 0 {
		return
	}
	s.selectedRow += delta
	if s.selectedRow < 0 {
		s.selectedRow = len(s.view.Rows) - 1
	}
	if s.selectedRow >= len(s.view.Rows) {
		s.selectedRow = 0
	}
}

func (s *Session) moveColumn(delta int) {
	if len(s.view.Columns) == 0 {
		return
	}
	s.selectedColumn += delta
	if s.selectedColumn < 0 {
		s.selectedColumn = len(s.view.Columns) - 1
	}
	if s.selectedColumn >= len(s.view.Columns) {
		s.selectedColumn = 0
	}
}

func (s *Session) toggleMark() {
	ids := s.selectedIDsFromCurrentRow()
	if len(ids) == 0 {
		return
	}
	id := ids[0]
	if _, ok := s.marks[id]; ok {
		delete(s.marks, id)
		s.markAnchor = s.selectedRow
		return
	}
	s.marks[id] = struct{}{}
	s.markAnchor = s.selectedRow
}

func (s *Session) spanMark() {
	if len(s.view.IDs) == 0 {
		return
	}
	if s.markAnchor < 0 || s.markAnchor >= len(s.view.IDs) {
		s.toggleMark()
		return
	}
	start, end := s.markAnchor, s.selectedRow
	if start > end {
		start, end = end, start
	}
	for index := start; index <= end; index++ {
		s.marks[s.view.IDs[index]] = struct{}{}
	}
}

func (s *Session) clearMarks() {
	for id := range s.marks {
		delete(s.marks, id)
	}
	s.markAnchor = -1
}

func (s *Session) selectedIDsFromCurrentRow() []string {
	if s.selectedRow < 0 || s.selectedRow >= len(s.view.IDs) {
		return nil
	}
	return []string{s.view.IDs[s.selectedRow]}
}

func (s *Session) selectedRowContext() (string, int, error) {
	row := s.selectedRow
	if row < 0 || row >= len(s.view.IDs) {
		return "", 0, fmt.Errorf("%w: no selected rows", ErrInvalidAction)
	}
	return s.view.IDs[row], row, nil
}

func (s *Session) selectedIDs() []string {
	if len(s.marks) == 0 {
		return s.selectedIDsFromCurrentRow()
	}
	ids := make([]string, 0, len(s.marks))
	for _, id := range s.view.IDs {
		if _, ok := s.marks[id]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func (s *Session) sortBySelectedColumn() error {
	if s.selectedColumn < 0 || s.selectedColumn >= len(s.view.Columns) {
		return fmt.Errorf("%w: selected column out of range", ErrUnsupportedHotKey)
	}
	s.sortByColumn(s.view.Columns[s.selectedColumn], true)
	return nil
}

func (s *Session) sortByColumn(column string, defaultAsc bool) {
	index := findColumnIndex(s.view.Columns, column)
	if index < 0 {
		return
	}
	asc := defaultAsc
	if s.sortColumn == column {
		asc = !s.sortAsc
	}
	s.reorderRows(index, asc)
	s.sortColumn = column
	s.sortAsc = asc
	s.selectedColumn = index
}

func (s *Session) invertSort() error {
	if s.sortColumn == "" {
		return fmt.Errorf("%w: no active sort", ErrUnsupportedHotKey)
	}
	s.sortByColumn(s.sortColumn, s.sortAsc)
	return nil
}

func findColumnIndex(columns []string, name string) int {
	for index, column := range columns {
		if column == name {
			return index
		}
	}
	return -1
}

func (s *Session) reorderRows(columnIndex int, asc bool) {
	rows := make([]tableRow, 0, len(s.view.Rows))
	for index, row := range s.view.Rows {
		rows = append(rows, tableRow{id: s.view.IDs[index], cells: append([]string{}, row...)})
	}
	sort.SliceStable(rows, func(i int, j int) bool {
		return lessCell(rows[i].cells[columnIndex], rows[j].cells[columnIndex], asc)
	})
	for index, row := range rows {
		s.view.IDs[index] = row.id
		s.view.Rows[index] = row.cells
	}
	s.clampSelectedRow()
}

func (s *Session) clampSelectedRow() {
	if len(s.view.Rows) == 0 {
		s.selectedRow = 0
		return
	}
	if s.selectedRow >= len(s.view.Rows) {
		s.selectedRow = len(s.view.Rows) - 1
	}
	if s.selectedRow < 0 {
		s.selectedRow = 0
	}
}

func lessCell(left string, right string, asc bool) bool {
	leftInt, leftErr := strconv.Atoi(left)
	rightInt, rightErr := strconv.Atoi(right)
	if leftErr == nil && rightErr == nil {
		if asc {
			return leftInt < rightInt
		}
		return leftInt > rightInt
	}
	leftLower := strings.ToLower(left)
	rightLower := strings.ToLower(right)
	if asc {
		return leftLower < rightLower
	}
	return leftLower > rightLower
}

func containsAction(actions []string, action string) bool {
	for _, candidate := range actions {
		if candidate == action {
			return true
		}
	}
	return false
}

func filterView(view ResourceView, filter string) ResourceView {
	filtered := ResourceView{
		Resource:    view.Resource,
		Columns:     append([]string{}, view.Columns...),
		Rows:        make([][]string, 0, len(view.Rows)),
		IDs:         make([]string, 0, len(view.IDs)),
		SortHotKeys: view.SortHotKeys,
		Actions:     append([]string{}, view.Actions...),
	}
	for index, row := range view.Rows {
		if !rowMatchesFilter(row, filter) {
			continue
		}
		filtered.Rows = append(filtered.Rows, append([]string{}, row...))
		if index < len(view.IDs) {
			filtered.IDs = append(filtered.IDs, view.IDs[index])
		}
	}
	return filtered
}

func clampSelectionIndex(value int, length int) int {
	if length == 0 {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value >= length {
		return length - 1
	}
	return value
}

func rowMatchesFilter(row []string, filter string) bool {
	for _, value := range row {
		if strings.Contains(strings.ToLower(value), filter) {
			return true
		}
	}
	return false
}

func (s *Session) vmDetailsByID(id string) (ResourceDetails, error) {
	vm, ok := findVMRowByID(s.navigator.catalog.VMs, id)
	if !ok {
		return ResourceDetails{}, fmt.Errorf("%w: no selected rows", ErrInvalidAction)
	}
	return vmDetails(vm), nil
}

func findVMRowByID(rows []VMRow, id string) (VMRow, bool) {
	for _, row := range rows {
		if row.Name == id {
			return row, true
		}
	}
	return VMRow{}, false
}

func vmDetails(row VMRow) ResourceDetails {
	fields := []DetailField{
		{Key: "NAME", Value: row.Name},
		{Key: "POWER_STATE", Value: defaultCell(row.PowerState)},
		{Key: "CPU_COUNT", Value: strconv.Itoa(row.CPUCount)},
		{Key: "MEMORY_MB", Value: strconv.Itoa(row.MemoryMB)},
		{Key: "COMMENTS", Value: defaultCell(row.Comments)},
		{Key: "DESCRIPTION", Value: defaultCell(row.Description)},
		{Key: "SNAPSHOT_COUNT", Value: strconv.Itoa(vmSnapshotCount(row))},
	}
	fields = append(fields, vmSnapshotFields(row.Snapshots)...)
	return ResourceDetails{Title: "VM DETAILS", Fields: fields}
}

func vmSnapshotCount(row VMRow) int {
	if row.SnapshotCount > 0 {
		return row.SnapshotCount
	}
	return len(row.Snapshots)
}

func vmSnapshotFields(snapshots []VMSnapshot) []DetailField {
	fields := make([]DetailField, 0, len(snapshots))
	for index, snapshot := range snapshots {
		fields = append(fields, DetailField{
			Key:   fmt.Sprintf("SNAPSHOT_%d", index+1),
			Value: snapshotFieldValue(snapshot),
		})
	}
	return fields
}

func snapshotFieldValue(snapshot VMSnapshot) string {
	return fmt.Sprintf("%s @ %s", defaultCell(snapshot.Identifier), defaultCell(snapshot.Timestamp))
}

func genericDetailsFromRow(view ResourceView, rowIndex int) ResourceDetails {
	fields := make([]DetailField, 0, len(view.Columns))
	if rowIndex >= 0 && rowIndex < len(view.Rows) {
		fields = detailFieldsFromColumns(view.Columns, view.Rows[rowIndex])
	}
	return ResourceDetails{
		Title:  strings.ToUpper(string(view.Resource)) + " DETAILS",
		Fields: fields,
	}
}

func detailFieldsFromColumns(columns []string, row []string) []DetailField {
	fields := make([]DetailField, 0, len(columns))
	for index, key := range columns {
		value := "-"
		if index < len(row) {
			value = defaultCell(row[index])
		}
		fields = append(fields, DetailField{Key: key, Value: value})
	}
	return fields
}

func normalizeResourceName(name string) (Resource, bool) {
	resource, ok := resourceAliasMap[strings.ToLower(strings.TrimSpace(name))]
	return resource, ok
}

func (s *Session) applyStoredColumns(view ResourceView) (ResourceView, error) {
	stored, ok := s.columnSelection[view.Resource]
	if !ok {
		return view, nil
	}
	filteredView, _, err := selectVisibleColumns(view, stored)
	return filteredView, err
}

func selectVisibleColumns(view ResourceView, columns []string) (ResourceView, []string, error) {
	normalized := normalizeColumnSelection(columns)
	if len(normalized) == 0 {
		return ResourceView{}, nil, fmt.Errorf("%w: empty selection", ErrInvalidColumns)
	}
	indexByColumn := map[string]int{}
	for index, column := range view.Columns {
		indexByColumn[strings.ToUpper(column)] = index
	}
	indexes := make([]int, 0, len(normalized))
	for _, column := range normalized {
		index, ok := indexByColumn[column]
		if !ok {
			return ResourceView{}, nil, fmt.Errorf("%w: unknown column %s", ErrInvalidColumns, column)
		}
		indexes = append(indexes, index)
	}
	rows := make([][]string, len(view.Rows))
	for rowIndex, row := range view.Rows {
		rows[rowIndex] = visibleColumnsRow(row, indexes)
	}
	return ResourceView{
		Resource:    view.Resource,
		Columns:     resolvedColumns(view.Columns, indexes),
		Rows:        rows,
		IDs:         append([]string{}, view.IDs...),
		SortHotKeys: visibleSortHotKeys(view.SortHotKeys, resolvedColumns(view.Columns, indexes)),
		Actions:     append([]string{}, view.Actions...),
	}, normalized, nil
}

func visibleColumnsRow(row []string, indexes []int) []string {
	visible := make([]string, len(indexes))
	for index, columnIndex := range indexes {
		if columnIndex >= 0 && columnIndex < len(row) {
			visible[index] = row[columnIndex]
		}
	}
	return visible
}

func normalizeColumnSelection(columns []string) []string {
	normalized := make([]string, 0, len(columns))
	seen := map[string]struct{}{}
	for _, raw := range columns {
		value := strings.ToUpper(strings.TrimSpace(raw))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func resolvedColumns(all []string, indexes []int) []string {
	columns := make([]string, 0, len(indexes))
	for _, index := range indexes {
		if index >= 0 && index < len(all) {
			columns = append(columns, all[index])
		}
	}
	return columns
}

func visibleSortHotKeys(hotKeys map[string]string, visibleColumns []string) map[string]string {
	visible := map[string]struct{}{}
	for _, column := range visibleColumns {
		visible[column] = struct{}{}
	}
	filtered := map[string]string{}
	for key, column := range hotKeys {
		if _, ok := visible[column]; ok {
			filtered[key] = column
		}
	}
	return filtered
}

// ResourceCommandAliases returns all supported resource aliases as colon commands.
func ResourceCommandAliases() []string {
	aliases := make([]string, 0, len(resourceAliasMap))
	for alias := range resourceAliasMap {
		aliases = append(aliases, ":"+alias)
	}
	sort.Strings(aliases)
	return aliases
}

type tableRow struct {
	id    string
	cells []string
}

const defaultBodyViewportRows = 10

func columnWidths(columns []string, rows [][]string) []int {
	widths := make([]int, len(columns))
	for index, column := range columns {
		widths[index] = len(column)
	}
	for _, row := range rows {
		updateWidths(widths, row)
	}
	return widths
}

func updateWidths(widths []int, row []string) {
	for index, value := range row {
		if index < len(widths) && len(value) > widths[index] {
			widths[index] = len(value)
		}
	}
}

func formatCells(cells []string, widths []int) string {
	parts := make([]string, 0, len(cells))
	for index, value := range cells {
		if index >= len(widths) {
			parts = append(parts, value)
			continue
		}
		parts = append(parts, padRight(value, widths[index]))
	}
	return strings.Join(parts, "  ") + "\n"
}

func padRight(value string, width int) string {
	padding := width - len(value)
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}

func viewportRows(rows [][]string, selectedRow int) [][]string {
	if len(rows) <= defaultBodyViewportRows {
		return rows
	}
	start, end := viewportBounds(len(rows), selectedRow, defaultBodyViewportRows)
	return rows[start:end]
}

func viewportBounds(totalRows int, selectedRow int, maxRows int) (int, int) {
	if maxRows <= 0 || totalRows <= maxRows {
		return 0, totalRows
	}
	normalized := selectedRow
	if normalized < 0 {
		normalized = 0
	}
	if normalized >= totalRows {
		normalized = totalRows - 1
	}
	start := normalized - maxRows + 1
	if start < 0 {
		start = 0
	}
	return start, start + maxRows
}
