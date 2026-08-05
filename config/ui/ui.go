package ui

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"

	"github.com/hanzoai/status/storage"
	static "github.com/hanzoai/status/web"
)

const (
	defaultTitle               = "Status"
	defaultDescription         = "Automated status page for monitoring endpoints"
	defaultHeader              = "Status"
	defaultDashboardHeading    = ""
	defaultDashboardSubheading = ""
	defaultLink                = ""
	defaultCustomCSS           = ""
	defaultSortBy              = "name"
	defaultFilterBy            = "none"

	// brandRoot is where artwork lives, and the only place it lives.
	brandRoot = "/brands/"

	// The two colours the app paints (web/app/src/index.css --color-background:
	// hsl(0 0% 100%) and hsl(222.2 84% 4.9%)). The browser chrome and the PWA
	// splash read them from here so the first pixel and the second one match.
	backgroundLight = "#ffffff"
	backgroundDark  = "#020817"
)

var (
	defaultDarkMode = true

	ErrButtonValidationFailed = errors.New("invalid button configuration: missing required name or link")
	ErrInvalidDefaultSortBy   = errors.New("invalid default-sort-by value: must be 'name', 'group', or 'health'")
	ErrInvalidDefaultFilterBy = errors.New("invalid default-filter-by value: must be 'none', 'failing', or 'unstable'")
	ErrBrandAssetMissing      = errors.New("invalid brand: the named brand is missing an asset")
)

// Config is the configuration for the status page UI
type Config struct {
	Title                   string   `yaml:"title,omitempty"`                  // Title of the page
	Description             string   `yaml:"description,omitempty"`            // Meta description of the page
	DashboardHeading        string   `yaml:"dashboard-heading,omitempty"`      // Dashboard Title between header and endpoints
	DashboardSubheading     string   `yaml:"dashboard-subheading,omitempty"`   // Dashboard Description between header and endpoints
	Header                  string   `yaml:"header,omitempty"`                 // Header is the text at the top of the page
	Brand                   string   `yaml:"brand,omitempty"`                  // Brand is the directory under /brands whose mark this page wears
	Link                    string   `yaml:"link,omitempty"`                   // Link to open when clicking on the logo
	Buttons                 []Button `yaml:"buttons,omitempty"`                // Buttons to display below the header
	CustomCSS               string   `yaml:"custom-css,omitempty"`             // Custom CSS to include in the page
	DarkMode                *bool    `yaml:"dark-mode,omitempty"`              // DarkMode is a flag to enable dark mode by default
	DefaultSortBy           string   `yaml:"default-sort-by,omitempty"`        // DefaultSortBy is the default sort option ('name', 'group', 'health')
	DefaultFilterBy         string   `yaml:"default-filter-by,omitempty"`      // DefaultFilterBy is the default filter option ('none', 'failing', 'unstable')
	//////////////////////////////////////////////
	// Non-configurable - used for UI rendering //
	//////////////////////////////////////////////
	MaximumNumberOfResults int        `yaml:"-"` // MaximumNumberOfResults to display on the page, it's not configurable because we're passing it from the storage config
	Icons                  Icons      `yaml:"-"` // Icons are derived from Brand by ValidateAndSetDefaults
	Background             Background `yaml:"-"` // Background is the colour pair the app paints, not a choice a deployment makes
}

func (cfg *Config) IsDarkMode() bool {
	if cfg.DarkMode != nil {
		return *cfg.DarkMode
	}
	return defaultDarkMode
}

// Button is the configuration for a button on the UI
type Button struct {
	Name string `yaml:"name,omitempty"` // Name is the text to display on the button
	Link string `yaml:"link,omitempty"` // Link to open when the button is clicked.
}

// Validate validates the button configuration
func (btn *Button) Validate() error {
	if len(btn.Name) == 0 || len(btn.Link) == 0 {
		return ErrButtonValidationFailed
	}
	return nil
}

// Icons are every mark the page wears: the tab, the home screen, the installed
// app, and the header logo. They are derived from Brand rather than configured
// one by one, because one binary serves every brand we run and the marks used to
// be listed per-deployment. List them and a deployment can name Hanzo in its
// title and still hand out Lux's icon; derive them and it cannot.
type Icons struct {
	SVG      string // tab icon, and the one browsers prefer
	ICO      string // tab icon for browsers that ask for /favicon.ico and nothing else
	PNG16    string
	PNG32    string
	Touch180 string // iOS home screen
	PNG192   string // PWA install
	PNG512   string // PWA splash
	Logo     string // the wordmark in the page header
}

// iconsFor resolves every mark from the one brand directory. An unnamed brand
// gets empty strings and the page then renders with no mark at all, which is the
// honest answer for a deployment that never said who it is — and a better one
// than quietly wearing whichever brand's files happened to sit at the web root.
func iconsFor(brand string) Icons {
	if len(brand) == 0 {
		return Icons{}
	}
	dir := brandRoot + brand + "/"
	return Icons{
		SVG:      dir + "favicon.svg",
		ICO:      dir + "favicon.ico",
		PNG16:    dir + "favicon-16.png",
		PNG32:    dir + "favicon-32.png",
		Touch180: dir + "icon-180.png",
		PNG192:   dir + "icon-192.png",
		PNG512:   dir + "icon-512.png",
		Logo:     dir + "logo.svg",
	}
}

// validate proves every mark this brand promises is actually in the binary. A
// misspelled brand is otherwise invisible until someone notices the tab icon is
// missing — or worse, does not notice that it is another company's.
func (icons Icons) validate() error {
	if len(icons.SVG) == 0 { // no brand named, so there is nothing to prove
		return nil
	}
	for _, asset := range []string{icons.SVG, icons.ICO, icons.PNG16, icons.PNG32, icons.Touch180, icons.PNG192, icons.PNG512, icons.Logo} {
		file, err := static.FileSystem.Open(static.RootPath + asset)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrBrandAssetMissing, asset)
		}
		_ = file.Close()
	}
	return nil
}

// Background is the pair of colours the app actually paints, mirrored from
// web/app/src/index.css. The browser chrome reads both (it switches on
// prefers-color-scheme like the page does) and the PWA splash reads one.
type Background struct {
	Dark  string
	Light string
}

// BackgroundColor is the single colour a PWA splash gets, since a manifest can
// only hold one: whichever of the pair this deployment defaults to.
func (cfg *Config) BackgroundColor() string {
	if cfg.IsDarkMode() {
		return cfg.Background.Dark
	}
	return cfg.Background.Light
}

// GetDefaultConfig returns a Config struct with the default values
func GetDefaultConfig() *Config {
	return &Config{
		Title:                  defaultTitle,
		Description:            defaultDescription,
		DashboardHeading:       defaultDashboardHeading,
		DashboardSubheading:    defaultDashboardSubheading,
		Header:                 defaultHeader,
		Link:                   defaultLink,
		CustomCSS:              defaultCustomCSS,
		DarkMode:               &defaultDarkMode,
		DefaultSortBy:          defaultSortBy,
		DefaultFilterBy:        defaultFilterBy,
		MaximumNumberOfResults: storage.DefaultMaximumNumberOfResults,
		Background:             Background{Dark: backgroundDark, Light: backgroundLight},
	}
}

// ValidateAndSetDefaults validates the UI configuration and sets the default values if necessary.
func (cfg *Config) ValidateAndSetDefaults() error {
	if len(cfg.Title) == 0 {
		cfg.Title = defaultTitle
	}
	if len(cfg.Description) == 0 {
		cfg.Description = defaultDescription
	}
	if len(cfg.DashboardHeading) == 0 {
		cfg.DashboardHeading = defaultDashboardHeading
	}
	if len(cfg.DashboardSubheading) == 0 {
		cfg.DashboardSubheading = defaultDashboardSubheading
	}
	if len(cfg.Header) == 0 {
		cfg.Header = defaultHeader
	}
	if len(cfg.Link) == 0 {
		cfg.Link = defaultLink
	}
	if len(cfg.CustomCSS) == 0 {
		cfg.CustomCSS = defaultCustomCSS
	}
	if cfg.DarkMode == nil {
		cfg.DarkMode = &defaultDarkMode
	}
	if len(cfg.DefaultSortBy) == 0 {
		cfg.DefaultSortBy = defaultSortBy
	} else if cfg.DefaultSortBy != "name" && cfg.DefaultSortBy != "group" && cfg.DefaultSortBy != "health" {
		return ErrInvalidDefaultSortBy
	}
	if len(cfg.DefaultFilterBy) == 0 {
		cfg.DefaultFilterBy = defaultFilterBy
	} else if cfg.DefaultFilterBy != "none" && cfg.DefaultFilterBy != "failing" && cfg.DefaultFilterBy != "unstable" {
		return ErrInvalidDefaultFilterBy
	}
	cfg.Icons = iconsFor(cfg.Brand)
	if err := cfg.Icons.validate(); err != nil {
		return err
	}
	cfg.Background = Background{Dark: backgroundDark, Light: backgroundLight}
	for _, btn := range cfg.Buttons {
		if err := btn.Validate(); err != nil {
			return err
		}
	}
	// Validate that the template works
	t, err := template.ParseFS(static.FileSystem, static.IndexPath)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	return t.Execute(&buffer, ViewData{UI: cfg, Theme: "dark"})
}

type ViewData struct {
	UI    *Config
	Theme string
}
