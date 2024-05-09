import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { GenerateDiagram, LoadConfig, SaveConfig } from "../wailsjs/go/main/App"
import { OpenProjectDialog, SaveProjectDialog, OpenProject } from '../wailsjs/go/main/App'
import { FetchModel, UpdateModel, ImportModelDialog } from "../wailsjs/go/main/App"
import { FetchAnalysis, UpdateAnalysis, AddAnalysisCase, RemoveAnalysisCase } from "../wailsjs/go/main/App"
import { FetchEvaluate, UpdateEvaluate, SelectExec, EvaluateCase, GetEvaluateLog, CancelEvaluate } from "../wailsjs/go/main/App"
import { FetchResults, OpenCaseDirDialog } from "../wailsjs/go/main/App"
import { GetModeViz } from "../wailsjs/go/main/App"
// import { EvaluateLinearization, CancelEvaluate } from "../wailsjs/go/main/App"
import { main, diagram as diag, viz } from "../wailsjs/go/models"
import { EventsOn } from "../wailsjs/runtime/runtime"
import { LogError } from '../wailsjs/runtime/runtime'

export const NOT_LOADED = 0;
export const LOADING = 1;
export const LOADED = 2;

// Contains loading status of the various project components
class Loading {
    project: number = 0;
    config: number = 0;
    info: number = 0;
    model: number = 0;
    analysis: number = 0;
    evaluate: number = 0;
    results: number = 0;
    diagram: number = 0;
    viz: number = 0;
}



export const useProjectStore = defineStore('project', () => {

    const saving = ref(false)
    const loaded = ref(false)
    const config = reactive<main.Config>(new main.Config)
    const info = reactive<main.Info>(new main.Info)
    const model = reactive<main.Model>(new main.Model)
    const analysis = reactive<main.Analysis>(new main.Analysis)
    const evaluate = reactive<main.Evaluate>(new main.Evaluate)
    const results = reactive<main.Results>(new main.Results)
    const evalStatus = reactive<Array<main.EvalStatus>>(new Array)
    const evalCaseID = ref(1)
    const diagram = reactive<diag.Diagram>(new diag.Diagram)
    const status = reactive<Loading>(new Loading)
    const modeViz = reactive<Array<viz.ModeData>>(new Array)

    // Load config when store is initialized
    LoadConfig().then(result => {
        Object.assign(config, result)
    }).catch(err => {
        LogError(err)
        console.log(err)
    })

    function $reset() {
        loaded.value = false
    }

    //--------------------------------------------------------------------------
    // Project
    //--------------------------------------------------------------------------

    function openDialog() {
        OpenProjectDialog().then(result => {
            Object.assign(info, result)
            updateRecentProjects(info.Path)
            loaded.value = true
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    function saveDialog() {
        SaveProjectDialog().then(result => {
            Object.assign(info, result)
            updateRecentProjects(info.Path)
            loaded.value = true
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    function open(path: string) {
        OpenProject(path).then(result => {
            Object.assign(info, result)
            updateRecentProjects(info.Path)
            loaded.value = true
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    function updateRecentProjects(path: string) {
        // If path in recent, remove it
        const index = config.RecentProjects.indexOf(path)
        if (index > -1) config.RecentProjects.splice(index, 1)

        // Prepend new path
        config.RecentProjects.unshift(path)

        // Limit to 5 items
        config.RecentProjects = config.RecentProjects.slice(0, 5)
        // Save config
        SaveConfig(config).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    //--------------------------------------------------------------------------
    // Model
    //--------------------------------------------------------------------------

    function fetchModel() {
        return new Promise<main.Model>((resolve, reject) => {
            FetchModel().then(result => {
                Object.assign(model, result)
                resolve(model)
            }).catch(err => {
                LogError(err)
                console.log(err)
                reject(err)
            })
        })
    }

    function importModelDialog() {
        ImportModelDialog().then(result => {
            Object.assign(model, result)
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    function updateModel() {
        UpdateModel(model).then(result => {
            Object.assign(model, result)
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    //--------------------------------------------------------------------------
    // Analysis
    //--------------------------------------------------------------------------

    function fetchAnalysis() {
        return new Promise<main.Analysis>((resolve, reject) => {
            FetchAnalysis().then(result => {
                Object.assign(analysis, result)
                resolve(analysis)
            }).catch(err => {
                LogError(err)
                console.log(err)
                reject(err)
            })
        })
    }

    function updateAnalysis() {
        UpdateAnalysis(analysis).then(result => {
            Object.assign(analysis, result)
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    function addAnalysisCase() {
        return new Promise<main.Case>((resolve, reject) => {
            AddAnalysisCase().then(result => {
                Object.assign(analysis, result)
                resolve(result.Cases[result.Cases.length - 1])
            }).catch(err => {
                LogError(err)
                console.log(err)
                reject(err)
            })
        })
    }

    function removeAnalysisCase(id: number) {
        RemoveAnalysisCase(id).then(result => {
            Object.assign(analysis, result)
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    //--------------------------------------------------------------------------
    // Evaluate
    //--------------------------------------------------------------------------

    function fetchEvaluate() {
        return new Promise<main.Evaluate>((resolve, reject) => {
            FetchEvaluate().then(result => {
                Object.assign(evaluate, result)
                resolve(evaluate)
            }).catch(err => {
                LogError(err)
                console.log(err)
                reject(err)
            })
        })
    }

    function updateEvaluate() {
        UpdateEvaluate(evaluate).then(result => {
            Object.assign(evaluate, result)
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    function selectExec() {
        SelectExec().then(result => {
            Object.assign(evaluate, result)
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    // Setup listener for evaluation status updates
    EventsOn("evalStatus", (status: main.EvalStatus) => {
        Object.assign(evalStatus[status.ID - 1], status)
    })

    function startEvaluate(caseID: number) {
        EvaluateCase(caseID).then(result => {
            Object.assign(evalStatus, result)
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    function cancelEvaluate() {
        CancelEvaluate().catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    //--------------------------------------------------------------------------
    // Results
    //--------------------------------------------------------------------------

    function fetchResults() {
        FetchResults().then(result => {
            Object.assign(results, result)
            console.log(result)
        }).catch(err => {
            LogError(err)
            console.log(err)
        })
    }

    function openCaseDirDialog() {
        status.results = LOADING
        return new Promise<main.Results>((resolve, reject) => {
            OpenCaseDirDialog().then(result => {
                Object.assign(results, result)
                status.diagram = NOT_LOADED
                status.results = LOADED
                resolve(results)
            }).catch(err => {
                LogError(err)
                console.log(err)
                status.results = NOT_LOADED
                reject(err)
            })
        })
    }

    //--------------------------------------------------------------------------
    // Diagram
    //--------------------------------------------------------------------------

    function generateDiagram(doCluster: boolean) {
        status.diagram = LOADING
        return new Promise<diag.Diagram>((resolve, reject) => {
            GenerateDiagram(results.MinFreq, results.MaxFreq, doCluster).then(result => {
                Object.assign(diagram, result)
                status.diagram = LOADED
                resolve(diagram)
            }).catch(err => {
                LogError(err)
                console.log(err)
                status.diagram = NOT_LOADED
                reject(err)
            })
        })
    }

    //--------------------------------------------------------------------------
    // Visualization
    //--------------------------------------------------------------------------

    function getModeViz(opID: number, modeID: number, scale: number) {
        status.viz = LOADING
        return new Promise<viz.ModeData>((resolve, reject) => {
            GetModeViz(opID, modeID, scale).then(result => {
                console.log('result: ', result)
                const found = modeViz.find((md) => result.OPID == md.OPID && result.ModeID == md.ModeID)
                console.log('found: ', found)
                if (found !== undefined) {
                    Object.assign(found, result)
                } else {
                    modeViz.push(result);
                }
                console.log('modeViz: ', modeViz)
                status.viz = LOADED
                resolve(result)
            }).catch(err => {
                LogError(err)
                console.log(err)
                status.viz = NOT_LOADED
                reject(err)
            })
        })
    }

    function clearModeViz() {
        modeViz.splice(0);
    }

    //--------------------------------------------------------------------------
    // Other
    //--------------------------------------------------------------------------

    return {
        status: status,
        loaded,
        saving,
        $reset,
        // Project
        config,
        info,
        saveDialog,
        openDialog,
        open,
        // Model
        model,
        fetchModel,
        importModelDialog,
        updateModel,
        // Analysis
        analysis,
        fetchAnalysis,
        updateAnalysis,
        addAnalysisCase,
        removeAnalysisCase,
        // Evaluate
        evaluate,
        evalStatus,
        fetchEvaluate,
        updateEvaluate,
        startEvaluate,
        cancelEvaluate,
        selectExec,
        evalCaseID,
        // Results
        results,
        fetchResults,
        openCaseDirDialog,
        // Diagram
        diagram,
        generateDiagram,
        // Visualization
        getModeViz,
        clearModeViz,
        modeViz,
    }
})
