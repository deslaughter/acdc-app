package main

import (
	"acdc/diagram"
	"acdc/viz"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sync/errgroup"
)

const ProjectFile = "project.json"

// App struct
type App struct {
	ctx     context.Context
	Project *Project
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

//------------------------------------------------------------------------------
// Project
//------------------------------------------------------------------------------

func (a *App) OpenProjectDialog() (*Info, error) {

	// Open dialog so user can select the file
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Project",
		Filters: []runtime.FileFilter{
			{DisplayName: "Projects (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error selecting project file: %w", err)
	}

	// If path not selected, return
	if path == "" {
		return nil, fmt.Errorf("no file selected")
	}

	// Load project
	a.Project, err = LoadProject(path)
	if err != nil {
		return nil, err
	}

	// Set project path
	a.Project.Info.Path = path

	// set window title
	runtime.WindowSetTitle(a.ctx, "ACDC - "+path)

	// Open project
	return &a.Project.Info, err
}

// SaveProjectDialog opens a dialog to select where to save the project
func (a *App) SaveProjectDialog() (*Info, error) {

	// Open dialog so user can select the file
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Project As",
		DefaultFilename: "project.json",
		Filters: []runtime.FileFilter{
			{DisplayName: "Projects (*.json)", Pattern: "*.json"},
		},
		CanCreateDirectories: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error selecting project directory: %w", err)
	}

	// Create path
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, fmt.Errorf("error creating project directory '%s': %w", path, err)
	}

	// If project not loaded, create new project
	if a.Project == nil {
		a.Project = NewProject()
	}

	// Set project path
	a.Project.Info.Path = path

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	// Return full project
	return &a.Project.Info, nil
}

func (a *App) OpenProject(path string) (*Info, error) {

	// Load project
	var err error
	a.Project, err = LoadProject(path)
	if err != nil {
		return nil, err
	}

	// Set project path
	a.Project.Info.Path = path

	return &a.Project.Info, nil
}

//------------------------------------------------------------------------------
// Model
//------------------------------------------------------------------------------

// UpdateModel saves changes to the current project
func (a *App) FetchModel() (*Model, error) {

	// If no Model in project, create it
	if a.Project.Model == nil {
		a.Project.Model = NewModel()
	}

	return a.Project.Model, nil
}

// UpdateModel saves changes to the current project
func (a *App) UpdateModel(model *Model) (*Model, error) {

	// Update model in project
	a.Project.Model = model

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	return a.Project.Model, nil
}

func (a *App) ImportModelDialog() (*Model, error) {

	// Open dialog so user can select the file
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Model",
		Filters: []runtime.FileFilter{
			{DisplayName: "OpenFAST Model (*.fst)", Pattern: "*.fst"},
		},
	})
	if err != nil {
		// No file was selected
		return a.Project.Model, nil
	}

	// Parse model files
	files, err := ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("error importing OpenFAST model '%s': %w", path, err)
	}

	// Get imported paths
	paths := []string{}
	for p := range files.PathMap {
		p, _ := filepath.Rel(filepath.Dir(path), p)
		paths = append(paths, p)
	}
	sort.Strings(paths)

	// Initialize models structure
	a.Project.Model = &Model{
		HasAero: ((len(files.AeroDyn)+len(files.AeroDyn14)) > 0 &&
			len(files.InflowWind) > 0),
		Files:         files,
		ImportedPaths: paths,
		Notes:         []string{},
	}

	// Save project
	if _, err := a.Project.Save(); err != nil {
		runtime.LogErrorf(a.ctx, "SelectExec: error saving project: %s", err)
	}

	// Parse and return model
	return a.Project.Model, nil
}

//------------------------------------------------------------------------------
// Analysis
//------------------------------------------------------------------------------

func (a *App) FetchAnalysis() (*Analysis, error) {

	// If no analysis in project, create it
	if a.Project.Analysis == nil {
		a.Project.Analysis = NewAnalysis()
	}

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	return a.Project.Analysis, nil
}

// UpdateEval runs Case.Calculate and saves evaluation data
func (a *App) UpdateAnalysis(an *Analysis) (*Analysis, error) {

	// Calculate analysis cases
	if err := an.Calculate(); err != nil {
		return nil, err
	}

	// Update analysis in the project
	a.Project.Analysis = an

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	return a.Project.Analysis, nil
}

func (a *App) AddAnalysisCase() (*Analysis, error) {

	// If no analysis in project, create it
	if a.Project.Analysis == nil {
		a.Project.Analysis = NewAnalysis()
	}

	// Add new case to analysis
	a.Project.Analysis.Cases = append(a.Project.Analysis.Cases, NewCase())

	// Calculate analysis cases
	if err := a.Project.Analysis.Calculate(); err != nil {
		return nil, err
	}

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	return a.Project.Analysis, nil
}

func (a *App) RemoveAnalysisCase(ID int) (*Analysis, error) {

	// If no analysis in project, create it
	if a.Project.Analysis == nil {
		a.Project.Analysis = NewAnalysis()
	}

	// Filter out case that matches ID
	tmp := []Case{}
	for _, c := range a.Project.Analysis.Cases {
		if c.ID != ID {
			tmp = append(tmp, c)
		}
	}
	a.Project.Analysis.Cases = tmp

	// Calculate analysis cases
	if err := a.Project.Analysis.Calculate(); err != nil {
		return nil, err
	}

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	return a.Project.Analysis, nil
}

//------------------------------------------------------------------------------
// Evaluate
//------------------------------------------------------------------------------

func (a *App) FetchEvaluate() (*Evaluate, error) {

	// If no Eval in project, create it
	if a.Project.Evaluate == nil {
		a.Project.Evaluate = NewEvaluate()
	}

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	return a.Project.Evaluate, nil
}

// UpdateEval runs Case.Calculate and saves evaluation data
func (a *App) UpdateEvaluate(ev *Evaluate) (*Evaluate, error) {

	// Update analysis in the project
	a.Project.Evaluate = ev

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	return a.Project.Evaluate, nil
}

// SelectExec opens a dialog for the user to select an OpenFAST executable.
func (a *App) SelectExec() (*Evaluate, error) {

	// If no Eval in project, create it
	if a.Project.Evaluate == nil {
		a.Project.Evaluate = NewEvaluate()
	}

	// Lookup OpenFAST executable with default name
	execPath, err := exec.LookPath("openfast")
	if err != nil {
		execPath = ""
	}

	// Open dialog so user can select the executable path
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Select OpenFAST Executable",
		CanCreateDirectories: false,
		DefaultDirectory:     execPath,
	})
	if err != nil {
		return nil, err
	}

	// If no path selected, return current exec
	if path == "" {
		return a.Project.Evaluate, nil
	}

	// Get output from running the command
	output, _ := exec.Command(path).CombinedOutput()

	// If OpenFAST isn't in the output, return error
	if !bytes.Contains(output, []byte("OpenFAST")) {
		return nil, fmt.Errorf("'%s' is not an OpenFAST executable", path)
	}

	// Trim output
	if index := bytes.Index(output, []byte("Execution Info:")); index > -1 {
		output = output[:index-1]
	}
	if index := bytes.LastIndex(output, []byte("****")); index > -1 {
		output = output[index+4:]
	}

	// Update executable info in project
	a.Project.Evaluate.ExecPath = path
	a.Project.Evaluate.ExecVersion = string(bytes.TrimSpace(output))
	a.Project.Evaluate.ExecValid = true

	// Save project
	if _, err := a.Project.Save(); err != nil {
		runtime.LogErrorf(a.ctx, "SelectExec: error saving project: %s", err)
	}

	// Save project and return path
	return a.Project.Evaluate, nil
}

func (a *App) EvaluateCase(ID int) ([]EvalStatus, error) {

	// Clear results
	a.Project.Results = nil

	// Get case
	var c *Case
	for _, cc := range a.Project.Analysis.Cases {
		if cc.ID == ID {
			c = &cc
			break
		}
	}
	if c == nil {
		return nil, fmt.Errorf("Case ID %d not found", ID)
	}

	// Call existing cancel func
	EvalCancel(fmt.Errorf("new evaluation started"))

	// Create path to case directory
	caseDir := filepath.Join(strings.TrimSuffix(a.Project.Info.Path, filepath.Ext(a.Project.Info.Path)), fmt.Sprintf("case%02d", c.ID))
	if err := os.MkdirAll(caseDir, 0777); err != nil {
		return nil, fmt.Errorf("error creating directory '%s': %w", caseDir, err)
	}

	// Remove existing linearization files
	linFiles, err := filepath.Glob(filepath.Join(caseDir, "*.lin"))
	if err != nil {
		return nil, err
	}
	for _, linFile := range linFiles {
		os.Remove(linFile)
	}

	// Wrap app context with cancel function
	ctx, cancelFunc := context.WithCancelCause(a.ctx)

	// Save cancel function so it can be called
	EvalCancel = cancelFunc

	// Wrap cancel context with error group so eval will stop on first error
	g, ctx2 := errgroup.WithContext(ctx)

	// Create eval status slice
	statuses := []EvalStatus{}

	// Launch evaluations throttled to number of CPUs specified
	semChan := make(chan struct{}, a.Project.Evaluate.NumCPUs)
	for _, op := range c.OperatingPoints {
		op := op
		statuses = append(statuses, EvalStatus{ID: op.ID, State: "Queued"})
		g.Go(func() error {
			<-semChan
			defer func() { semChan <- struct{}{} }()
			return RunEvaluation(ctx2, a.Project.Model, c, &op, caseDir,
				a.Project.Evaluate.ExecPath)
		})
	}

	// Wait for evaluations to complete. If error, print
	go func() {
		if err := g.Wait(); err != nil {
			runtime.LogErrorf(a.ctx, "error evaluating case: %s", err)
		}
		close(semChan)
		for {
			if _, ok := <-semChan; !ok {
				break
			}
		}
		cancelFunc(nil) // cancel the context
	}()

	// Start evaluations
	go func() {
		time.Sleep(time.Second)
		for i := 0; i < a.Project.Evaluate.NumCPUs; i++ {
			semChan <- struct{}{}
		}
	}()

	return statuses, nil
}

func (a *App) CancelEvaluate() {
	EvalCancel(fmt.Errorf("evaluation canceled"))
}

func (a *App) GetEvaluateLog(path string) (string, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

//------------------------------------------------------------------------------
// Results
//------------------------------------------------------------------------------

func (a *App) FetchResults() (*Results, error) {

	// If no Results in project, create it
	if a.Project.Results == nil {
		a.Project.Results = NewResults()
	}

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	return a.Project.Results.ForApp(), nil
}

func (a *App) OpenCaseDirDialog() (*Results, error) {

	// Get path to project, if it doesn't exist, set to empty string
	projectDir := filepath.Dir(a.Project.Info.Path)
	if _, err := os.Stat(projectDir); err != nil {
		projectDir = ""
	}

	// Open dialog so user can select the case directory
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Open Case Directory",
		DefaultDirectory: projectDir,
	})
	if err != nil {
		return a.Project.Results, nil
	}

	// Process case directory to get results
	results, err := ProcessCaseDir(path)
	if err != nil {
		return nil, err
	}

	a.Project.Results = results

	// Save project
	if _, err := a.Project.Save(); err != nil {
		return nil, err
	}

	return a.Project.Results.ForApp(), nil
}

//--------------------------------------------------------------------------
// Diagram
//--------------------------------------------------------------------------

func (a *App) GenerateDiagram(maxFreqHz float64, doCluster bool) (*diagram.Diagram, error) {

	// Check that results have been loaded
	if a.Project.Results == nil {
		return nil, fmt.Errorf("load results before generating diagram")
	}

	// Generate diagram with given options
	diag, err := diagram.New(a.Project.Results.LinOPs, maxFreqHz, doCluster)
	if err != nil {
		return nil, err
	}

	return diag, nil
}

//--------------------------------------------------------------------------
// VTK
//--------------------------------------------------------------------------

func (a *App) GetVTKFile() ([]string, error) {
	// Read Turbine component VTK Files

	// Current working directory:  /Users/sryu/Desktop/FY24/ACDC/acdc-app
	curr_dir, err := os.Getwd() // get the current working directory
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Current working directory: ", curr_dir)

	// Glob *.vtp files
	pattern := "*.vtp"
	matches, err := os.ReadDir("viz/testdata")
	if err != nil {
		fmt.Println(err)
	}

	var vtp_files []string
	for _, match := range matches {
		f, _ := filepath.Match(pattern, match.Name())
		if f {
			vtp_files = append(vtp_files, match.Name())
		}
	}
	fmt.Println("Found vtp files: ", vtp_files)

	var out_list []string
	var empty_list []string
	for _, file := range vtp_files {
		if !(file == "BD_BldMotion1.001.vtp" || file == "ED_TowerLn2Mesh_motion.001.vtp") {
			continue
		}

		fmt.Println("Add file: ", file)
		out, err := os.ReadFile(curr_dir + "/viz/testdata/" + file)
		if err != nil {
			fmt.Println(err)
			return empty_list, err
		}
		out_list = append(out_list, "data:text/plain;base64,"+base64.StdEncoding.EncodeToString(out))
	}
	fmt.Println("out_list: ", len(out_list))

	return out_list, nil

	//out, err := os.ReadFile("/Users/sryu/Desktop/FY24/ACDC/acdc-app/viz/testdata/GroundSurface.vtp") // working
	// out, err := os.ReadFile("/Users/sryu/Desktop/FY24/ACDC/acdc-app/viz/testdata/cube_ascii.vtp") // working
	// out, err := os.ReadFile("/Users/sryu/Desktop/FY24/ACDC/acdc-app/viz/testdata/ED_Hub.001.vtp")
	//out, err := os.ReadFile("/Users/sryu/Desktop/FY24/ACDC/acdc-app/viz/testdata/BD_BldMotion1.001.vtp") // working
	// out, err := os.ReadFile(curr_dir + "/viz/testdata/ED_TowerLn2Mesh_motion.001.vtp") // working
	// if err != nil {
	// 	return "", err
	// }

	// return "data:text/plain;base64," + base64.StdEncoding.EncodeToString(out), nil
}

func (a *App) LoadVTKFile() ([]*viz.VTKFile, error) {
	testFiles := []string{
		"viz/testdata/BD_BldMotion1.001.vtp",
		"viz/testdata/ED_Hub.001.vtp",
		"viz/testdata/ED_Nacelle.001.vtp",
		"viz/testdata/ED_TailFin.001.vtp",
		"viz/testdata/ED_TowerLn2Mesh_motion.001.vtp",
		"viz/testdata/GroundSurface.vtp",
	}

	var resList []*viz.VTKFile

	for _, testFile := range testFiles {
		res, err := viz.LoadVTK(testFile)
		resList = append(resList, res)
		fmt.Println(res)
		if err != nil {
			fmt.Println(err)
		}
	}

	return resList, nil
}
