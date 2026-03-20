// Package web implementa o servidor de templates e assets estaticos.
package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// TemplateData contém os dados passados para os templates.
type TemplateData struct {
	Title             string
	PageTitle         string
	PageSubtitle      string
	ActivePage        string
	ShowLayout        bool
	UserInitials      string
	UserName          string
	TenantName        string
	NotificationCount int
	Stats             *DashboardStatsView
	Error             string
	Success           string
	Data              any
}

// DashboardStatsView é a view model do dashboard.
type DashboardStatsView struct {
	Revenue            RevenueView
	MechanicOS         MechanicOSView
	BakeryProducts     int
	Appointments       int
	PendingSuggestions int
}

type RevenueView struct {
	TodayFmt string
	WeekFmt  string
	Chart    []ChartDayView
	ByModule []ModuleView
}

type ChartDayView struct {
	Day            string
	Revenue        float64
	Tax            float64
	RevenuePercent float64
	TaxPercent     float64
}

type ModuleView struct {
	Module  string
	Count   int
	Percent float64
}

type MechanicOSView struct {
	Total         int
	Open          int
	InProgress    int
	AwaitApproval int
	Done          int
	OpenCount     int
	OpenPct       float64
	InProgressPct float64
	AwaitPct      float64
	DonePct       float64
}

// TemplateRenderer gerencia os templates HTML.
type TemplateRenderer struct {
	templates map[string]*template.Template
	baseDir   string
}

// NewTemplateRenderer cria um novo renderer de templates.
func NewTemplateRenderer(baseDir string) (*TemplateRenderer, error) {
	tr := &TemplateRenderer{
		templates: make(map[string]*template.Template),
		baseDir:   baseDir,
	}

	if err := tr.loadTemplates(); err != nil {
		return nil, fmt.Errorf("falha ao carregar templates: %w", err)
	}

	return tr, nil
}

func (tr *TemplateRenderer) loadTemplates() error {
	// Carrega o layout base
	layoutFile := filepath.Join(tr.baseDir, "layouts", "base.html")
	
	// Carrega todos os partials
	partialsGlob := filepath.Join(tr.baseDir, "partials", "*.html")
	partials, err := filepath.Glob(partialsGlob)
	if err != nil {
		return err
	}

	// Carrega cada página
	pagesGlob := filepath.Join(tr.baseDir, "pages", "*.html")
	pages, err := filepath.Glob(pagesGlob)
	if err != nil {
		return err
	}

	for _, page := range pages {
		name := strings.TrimSuffix(filepath.Base(page), ".html")
		
		// Combina: layout + partials + página
		files := []string{layoutFile}
		files = append(files, partials...)
		files = append(files, page)

		tmpl, err := template.ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("erro ao parsear template %s: %w", name, err)
		}

		tr.templates[name] = tmpl
	}

	return nil
}

// Render renderiza um template com os dados fornecidos.
func (tr *TemplateRenderer) Render(w http.ResponseWriter, name string, data TemplateData) error {
	tmpl, ok := tr.templates[name]
	if !ok {
		return fmt.Errorf("template %s nao encontrado", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}

// Reload recarrega todos os templates (útil em desenvolvimento).
func (tr *TemplateRenderer) Reload() error {
	return tr.loadTemplates()
}

// EmbeddedTemplateRenderer usa templates embedados no binário.
type EmbeddedTemplateRenderer struct {
	templates map[string]*template.Template
	fs        embed.FS
}

// NewEmbeddedTemplateRenderer cria um renderer com templates embedados.
func NewEmbeddedTemplateRenderer(efs embed.FS, baseDir string) (*EmbeddedTemplateRenderer, error) {
	tr := &EmbeddedTemplateRenderer{
		templates: make(map[string]*template.Template),
		fs:        efs,
	}

	if err := tr.loadTemplates(baseDir); err != nil {
		return nil, err
	}

	return tr, nil
}

func (tr *EmbeddedTemplateRenderer) loadTemplates(baseDir string) error {
	// Lê o layout
	layoutContent, err := fs.ReadFile(tr.fs, baseDir+"/layouts/base.html")
	if err != nil {
		return err
	}

	// Lê os partials
	partialFiles, err := fs.Glob(tr.fs, baseDir+"/partials/*.html")
	if err != nil {
		return err
	}

	var partialsContent strings.Builder
	for _, pf := range partialFiles {
		content, err := fs.ReadFile(tr.fs, pf)
		if err != nil {
			return err
		}
		partialsContent.Write(content)
		partialsContent.WriteString("\n")
	}

	// Lê e processa cada página
	pageFiles, err := fs.Glob(tr.fs, baseDir+"/pages/*.html")
	if err != nil {
		return err
	}

	for _, pf := range pageFiles {
		name := strings.TrimSuffix(filepath.Base(pf), ".html")
		pageContent, err := fs.ReadFile(tr.fs, pf)
		if err != nil {
			return err
		}

		// Combina tudo em um único template
		combined := string(layoutContent) + "\n" + partialsContent.String() + "\n" + string(pageContent)
		
		tmpl, err := template.New(name).Parse(combined)
		if err != nil {
			return fmt.Errorf("erro ao parsear template %s: %w", name, err)
		}

		tr.templates[name] = tmpl
	}

	return nil
}

// Render renderiza um template embedado.
func (tr *EmbeddedTemplateRenderer) Render(w http.ResponseWriter, name string, data TemplateData) error {
	tmpl, ok := tr.templates[name]
	if !ok {
		return fmt.Errorf("template %s nao encontrado", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}
