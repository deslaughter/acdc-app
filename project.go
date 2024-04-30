package main

import (
	"acdc/diagram"
	"acdc/viz"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Project struct {
	Info     Info             `json:"Info"`
	Model    *Model           `json:"Model"`
	Analysis *Analysis        `json:"Analysis"`
	Evaluate *Evaluate        `json:"Evaluate"`
	Results  *Results         `json:"Results"`
	Diagram  *diagram.Diagram `json:"Diagram"`
	VTKFile  *viz.VTKFile     `json:"VTKFile"`
}

type Info struct {
	Date string `json:"Date"`
	Path string `json:"Path"`
}

func NewProject() *Project {
	return &Project{
		Evaluate: NewEvaluate(),
		Model:    NewModel(),
	}
}

func LoadProject(path string) (*Project, error) {

	// Read project file
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading project: %w", err)
	}

	// Parse project file
	p := NewProject()
	if err := json.Unmarshal(bs, p); err != nil {
		return nil, fmt.Errorf("error parsing project: %w", err)
	}

	// Update project path
	p.Info.Path = path

	return p, nil
}

func (p *Project) Save() (*Project, error) {

	if p == nil {
		return nil, fmt.Errorf("project not loaded")
	}

	// Update time
	p.Info.Date = time.Now().Format(time.RFC850)

	// Create temporary project to save relevant parts
	pSave := Project{
		Info:     p.Info,
		Model:    p.Model,
		Analysis: p.Analysis,
		Evaluate: p.Evaluate,
	}

	// Convert project to json
	bs, err := json.MarshalIndent(pSave, "", "\t")
	if err != nil {
		return nil, fmt.Errorf("error encoding project: %w", err)
	}

	// Write project file
	err = os.WriteFile(p.Info.Path, bs, 0777)
	if err != nil {
		return nil, fmt.Errorf("error reading project: %w", err)
	}

	// If project results has a path, write file
	if p.Results != nil && p.Results.LinDir != "" {

		// Write results file
		bs, err := json.MarshalIndent(p.Results, "", "\t")
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(filepath.Join(p.Results.LinDir, "results.json"), bs, 0777)
		if err != nil {
			return nil, err
		}

		// If project diagram is not nil, write file
		if p.Diagram != nil {

			// Write Diagram file
			bs, err := json.MarshalIndent(p.Diagram, "", "\t")
			if err != nil {
				return nil, err
			}
			err = os.WriteFile(filepath.Join(p.Results.LinDir, "diagram.json"), bs, 0777)
			if err != nil {
				return nil, err
			}
		}
	}

	return p, nil
}
