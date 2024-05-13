// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';
import {diagram} from '../models';
import {viz} from '../models';

export function AddAnalysisCase():Promise<main.Analysis>;

export function CancelEvaluate():Promise<void>;

export function DuplicateAnalysisCase(arg1:number):Promise<main.Analysis>;

export function EvaluateCase(arg1:number):Promise<Array<main.EvalStatus>>;

export function ExportDiagramDataJSON(arg1:diagram.Diagram):Promise<void>;

export function FetchAnalysis():Promise<main.Analysis>;

export function FetchEvaluate():Promise<main.Evaluate>;

export function FetchModel():Promise<main.Model>;

export function FetchResults():Promise<main.Results>;

export function GenerateDiagram(arg1:diagram.Options):Promise<diagram.Diagram>;

export function GetEvaluateLog(arg1:string):Promise<string>;

export function GetModeViz(arg1:number,arg2:number,arg3:number):Promise<viz.ModeData>;

export function ImportAnalysisCaseCurve(arg1:number):Promise<main.Analysis>;

export function ImportModelDialog():Promise<main.Model>;

export function LoadConfig():Promise<main.Config>;

export function OpenProject(arg1:string):Promise<main.Info>;

export function OpenProjectDialog():Promise<main.Info>;

export function ProcessLinDir(arg1:string):Promise<main.Results>;

export function RemoveAnalysisCase(arg1:number):Promise<main.Analysis>;

export function SaveConfig(arg1:main.Config):Promise<void>;

export function SaveProjectDialog():Promise<main.Info>;

export function SelectCaseLinDir(arg1:number):Promise<main.LinDirData>;

export function SelectCustomLinDir():Promise<main.LinDirData>;

export function SelectExec():Promise<main.Evaluate>;

export function UpdateAnalysis(arg1:main.Analysis):Promise<main.Analysis>;

export function UpdateDiagram(arg1:diagram.Diagram):Promise<void>;

export function UpdateEvaluate(arg1:main.Evaluate):Promise<main.Evaluate>;

export function UpdateModel(arg1:main.Model):Promise<main.Model>;
