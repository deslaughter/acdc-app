<script lang="ts" setup>
import { h, ref, onMounted, computed, reactive, VueElement } from 'vue'
import { useProjectStore, LOADED, LOADING, NOT_LOADED } from '../project';
import { Scatter } from 'vue-chartjs'
import { Chart, ChartData, ChartOptions, ChartEvent, ActiveElement, ScriptableContext } from 'chart.js'
import { ChartComponentRef } from "vue-chartjs"
import { main, diagram, viz } from "../../wailsjs/go/models"
import chroma from 'chroma-js'
import ModeViz from "./ModeViz.vue"

const project = useProjectStore()

onMounted(() => {
    project.fetchResults()
})

const selectedOP = ref<main.OperatingPoint>()
const selectedLine = ref<diagram.Line | null>(null)
const selectedPoint = ref<diagram.Point | null>(null)
const freqChart = ref<ChartComponentRef<'scatter'> | null>(null)
const dampChart = ref<ChartComponentRef<'scatter'> | null>(null)
const doCluster = ref(false)
const xAxisWS = ref(true)
const rotorSpeedMods = [1, 3, 6, 9, 12, 15]
let clickedMVIdx = ref(0)

interface Graph {
    options: ChartOptions<'scatter'>
    data: ChartData<'scatter'>
}

function selectPoint(event: ChartEvent, elements: ActiveElement[], chart: Chart<"scatter">) {
    if (elements.length == 0) return
    if (elements[0].datasetIndex >= project.diagram.Lines.length) return;
    selectedLine.value = project.diagram.Lines[elements[0].datasetIndex];
    selectedPoint.value = selectedLine.value.Points[elements[0].index];
}

function selectLine(line: diagram.Line) {
    selectedLine.value = line;
    selectedPoint.value = null;
}

function getModeViz(opID: number, modeID: number) {
    project.getModeViz(opID, modeID, 15.0).then(result => {

    }).catch(err => {

    })
}

function selectModeViz(opID: number, modeID: number) {
    console.log('clicked op, mode:', opID, modeID)
    clickedMVIdx.value = project.modeViz.findIndex(x => (x.OPID == opID) && (x.ModeID == modeID))
    console.log('clickedMVIdx: ', clickedMVIdx)
}

const charts = computed(() => {

    const CD = project.diagram
    const xLabel = (xAxisWS && CD.HasWind) ? "Wind Speed (m/s)" : "Rotor Speed (RPM)"
    const xValues = (xAxisWS && CD.HasWind) ? CD.WindSpeeds : CD.RotSpeeds
    const freqMin = Math.min(...CD.Lines.filter(line => !line.Hidden).map(line => Math.min(...line.Points.map(p => p.NaturalFreqHz))))
    const freqMax = Math.max(...CD.Lines.filter(line => !line.Hidden).map(line => Math.max(...line.Points.map(p => p.NaturalFreqHz))))
    const dampMin = Math.min(...CD.Lines.filter(line => !line.Hidden).map(line => Math.min(...line.Points.map(p => p.DampingRatio))))
    const dampMax = Math.max(...CD.Lines.filter(line => !line.Hidden).map(line => Math.max(...line.Points.map(p => p.DampingRatio))))

    let objs = new Array<Graph>;

    const configs = [
        { label: "Natural Frequency (Hz)", isNatFreq: true },
        { label: "Damping Ratio (-)", isNatfreq: false },
    ]

    const lineColors = chroma.cubehelix().lightness([0.4, 0.75]).rotations(2).scale().colors(CD.Lines.length)

    for (const cfg of configs) {

        let data = { datasets: [] } as ChartData<'scatter'>

        // Set line color if not defined
        for (let i = 0; i < CD.Lines.length; i++) {
            if (CD.Lines[i].Color == "") {
                CD.Lines[i].Color = lineColors[i % lineColors.length]
            }
        }

        // Loop through mode sets
        data.datasets = data.datasets.concat(CD.Lines.map((line, i) => ({
            label: line.Label,
            data: line.Points.map(p => ({
                x: (xAxisWS && CD.HasWind) ? p.WindSpeed : p.RotSpeed,
                y: cfg.isNatFreq ? p.NaturalFreqHz : p.DampingRatio,
            })),
            borderColor: line.Color,
            showLine: true,
            hidden: line.Hidden,
        })))

        // Loop through rotor speed multipliers
        if (cfg.isNatFreq) {
            data.datasets = data.datasets.concat(rotorSpeedMods.map(rsm => ({
                label: rsm + 'P',
                data: CD.RotSpeeds.map((RotSpeed, i) => {
                    return {
                        x: xValues[i],
                        y: RotSpeed / 60 * rsm,
                    }
                }),
                pointStyle: false,
                showLine: true,
                borderDash: [4, 6],
                borderColor: "slategray",
                borderWidth: 2
            })))
        }

        // Plot selected point and visualization points if one is selected
        if (selectedPoint.value != null) {
            const p = selectedPoint.value;
            data.datasets.push({
                label: 'selectedPoint',
                data: [{
                    x: (xAxisWS && CD.HasWind) ? p.WindSpeed : p.RotSpeed,
                    y: cfg.isNatFreq ? p.NaturalFreqHz : p.DampingRatio,
                }],
                pointStyle: 'crossRot',
                borderColor: 'red',
                pointRadius: 12,
                pointHoverRadius: 12,
                pointBorderWidth: 5,
                pointHoverBorderWidth: 5,
            });
        }

        // Add data options to array of chart objects
        objs.push({
            data: data,
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false },
                    tooltip: {
                        filter: (e) => {
                            // Exclude selected point from tooltip
                            return e.dataset.label != "selectedPoint"
                        }
                    }
                },
                scales: {
                    x: {
                        title: { display: true, text: xLabel, font: { size: 18 } },
                        min: Math.min(...xValues),
                        max: Math.max(...xValues),
                        ticks: { font: { size: 16 } }
                    },
                    y: {
                        title: { display: true, text: cfg.label, font: { size: 18 } },
                        ticks: { font: { size: 16 } },
                        min: 0.95 * (cfg.isNatFreq ? freqMin : dampMin),
                        max: 1.05 * (cfg.isNatFreq ? freqMax : dampMax),
                    },
                },
                onClick: selectPoint,
                interaction: {
                    mode: 'nearest'
                },
                animation: { duration: 0 }
            } as ChartOptions<"scatter">,
        })
    }

    return objs
})


</script>

<template>
    <main>
        <div class="card mb-3">
            <div class="card-header hstack gap-3">
                <span>Linearization Data</span>
                <!-- <a class="btn btn-primary btn-sm me-3" @click="openResults">Open Results</a> -->
                <a class="btn btn-primary ms-auto" @click="project.openCaseDirDialog">Import Data</a>
            </div>

            <div v-if="project.status.results == LOADING" class="spinner-border text-primary my-3 mx-auto"
                role="status">
                <span class="visually-hidden">Loading...</span>
            </div>
            <div class="card-body" v-if="project.status.results == LOADED && project.results.LinDir">
                <div class="mb-3 row">
                    <label for="case-dir" class="col-sm-2 col-form-label">Directory</label>
                    <div class="col-sm-10">
                        <input type="text" readonly class="form-control-plaintext" id="case-dir"
                            :value="project.results.LinDir">
                    </div>
                </div>
                <div class="row">
                    <label for="inputPassword" class="col-sm-2 col-form-label">Operating Point</label>
                    <div class="col-sm-10">
                        <select class="form-select" v-model="selectedOP">
                            <option :value="null">None</option>
                            <option v-for="op in project.results.OPs" :value="op">
                                {{ op.ID + 1 }} -
                                {{ project.results.HasWind ? `${op.WindSpeed.toPrecision(3)} m/s` :
                    `${op.RotSpeed.toPrecision(3)} RPM` }}
                            </option>
                        </select>
                    </div>
                </div>
                <table class="table table-bordered mt-4 mb-0 text-center table-sm" v-if="selectedOP != null">
                    <thead>
                        <tr>
                            <th scope="col">Mode</th>
                            <th scope="col">Natural Frequency (Hz)</th>
                            <th scope="col">Damped Frequency (Hz)</th>
                            <th scope="col">Damping Ratio (-)</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="mode in selectedOP?.Modes">
                            <td>{{ mode.ID + 1 }}</td>
                            <td>{{ mode.NaturalFreqHz.toPrecision(5) }}</td>
                            <td>{{ mode.DampedFreqHz.toPrecision(5) }}</td>
                            <td>{{ mode.DampingRatio.toExponential(3) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div class="card mb-3" v-if="project.status.results == LOADED">
            <div class="card-header d-flex justify-content-between align-items-center">
                <span>Campbell Diagram</span>
            </div>
            <div class="card-body" v-if="project.results.LinDir">
                <form class="row g-3">
                    <div class="col-3">
                        <label for="minFreq" class="col-form-label">Min Frequency (Hz)</label>
                        <input type="text" class="form-control" id="minFreq" v-model.number="project.results.MinFreq">
                    </div>
                    <div class="col-3">
                        <label for="maxFreq" class="col-form-label">Max Frequency (Hz)</label>
                        <input type="text" class="form-control" id="maxFreq" v-model.number="project.results.MaxFreq">
                    </div>
                    <div class="col-3">
                        <label for="doCluster" class="col-form-label ms-auto">Spectral Clustering</label>
                        <select class="form-select" v-model="doCluster">
                            <option :value="true" selected>Enable</option>
                            <option :value="false">Disable</option>
                        </select>
                    </div>
                    <div class="col-12">
                        <a class=" btn btn-primary" @click="project.generateDiagram(doCluster)">Generate</a>
                    </div>
                </form>
            </div>

            <div class="card-body border-top" v-if="project.status.diagram == LOADING">
                <div style="width:100%;height: 65vh">
                    <div class="spinner-border text-primary align-middle mx-auto" role="status">
                        <span class="visually-hidden">Loading...</span>
                    </div>
                </div>
            </div>

            <div class="card-body border-top" v-if="project.status.diagram == LOADED">
                <div class="row">
                    <div class="col-sm-12 col-lg-6">
                        <div style="position: relative; height: 65vh">
                            <Scatter ref="freqChart" :options="charts[0].options" :data="charts[0].data" />
                        </div>
                    </div>
                    <div class="col-sm-12 col-lg-6">
                        <div style="position: relative; height: 65vh">
                            <Scatter ref="dampChart" :options="charts[1].options" :data="charts[1].data" />
                        </div>
                    </div>
                    <div class="col-12">
                        <div class="row row-cols-auto g-2 mt-2">
                            <div class="col" v-for="(line, i) in project.diagram.Lines.filter((line) => !line.Hidden)">
                                <a class="btn btn-outline-dark" role="button" @click="selectLine(line)">
                                    <div style="position: relative">
                                        <div style="display: block; height: 4px; width: 24px; position: absolute; top: 49%; left: 0"
                                            :style="{ 'background-color': line.Color }">
                                        </div>
                                        <div style="margin-left: 32px;">{{ line.Label }}</div>
                                    </div>
                                </a>
                            </div>
                        </div>
                    </div>
                    <div class="col-12">
                        <div class="row row-cols-auto g-2 mt-3">
                            <div class="col">
                                <label class="col-form-label">Hidden:</label>
                            </div>
                            <div class="col" v-for="(line, i) in project.diagram.Lines.filter((line) => line.Hidden)">
                                <a class="btn btn-outline-dark" role="button" @click="selectLine(line)">
                                    <div style="position: relative">
                                        <div style="display: block; height: 4px; width: 24px; position: absolute; top: 49%; left: 0"
                                            :style="{ 'background-color': line.Color }">
                                        </div>
                                        <div style="margin-left: 32px;">{{ line.Label }}</div>
                                    </div>
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="row row-cols-2 mb-3">
            <div class="col">
                <div class="card h-100" v-if="selectedLine != null">
                    <div class="card-header hstack">
                        <span>Line</span>
                        <a class="btn btn-primary btn-sm ms-auto" @click="selectedLine.Hidden = !selectedLine.Hidden">
                            {{ selectedLine.Hidden ? "Show" : "Hide" }}
                        </a>
                    </div>
                    <div class="card-body">
                        <form class="row g-3">
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="lineNumber" class="col-form-label">Number</label>
                                <input type="text" class="form-control" id="lineNumber" :value="selectedLine.ID"
                                    disabled>
                            </div>
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="lineLabel" class="col-form-label">Label</label>
                                <input type="text" class="form-control" id="lineLabel" v-model="selectedLine.Label">
                            </div>
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="lineColor" class="col-form-label">Color</Label>
                                <input type="color" class="form-control form-control-color w-100" id="lineColor"
                                    v-model="selectedLine.Color">
                            </div>
                            <!-- <div class="col-3">
                                <label for="lineColor" class="col-form-label">Style</Label>
                                <select class="form-control" id="lineStyle" v-model="selectedLine.Dash">
                                    <option :value="null">-</option>
                                    <option :value="[2, 3]">--</option>
                                </select>
                            </div> -->
                            <div class="col-12" v-if="!selectedLine.Hidden">
                                <label for="linePoints" class="col-form-label">Points</Label>
                                <select class="form-select" id="linePoints" v-model="selectedPoint">
                                    <option v-for="point in selectedLine.Points" :value="point">OP: {{ point.OpPtID }},
                                        Rotor Speed: {{ point.RotSpeed.toFixed(2) }}, Wind Speed: {{
                    point.WindSpeed.toFixed(2) }}, Natural Frequency: {{
                    point.NaturalFreqHz.toFixed(3) }}
                                    </option>
                                </select>
                            </div>
                        </form>
                    </div>
                </div>
            </div>

            <div class="col">
                <div class="card h-100" v-if="selectedPoint != null">
                    <div class="card-header hstack">
                        <span>Mode</span>
                        <a class="btn btn-primary btn-sm ms-auto"
                            @click="getModeViz(selectedPoint.OpPtID, selectedPoint.ModeID)">
                            Visualize
                        </a>
                    </div>
                    <div class="card-body">
                        <form class="row row-cols-auto g-3">
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="modeOP" class="col-form-label">Operating Point</label>
                                <input type="text" class="form-control" id="modeOP"
                                    v-model.number="selectedPoint.OpPtID" disabled>
                            </div>
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="modeID" class="col-form-label">Mode ID</Label>
                                <input type="text" class="form-control" id="modeID"
                                    v-model.number="selectedPoint.ModeID" disabled>
                            </div>
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="modeRotSpeed" class="col-form-label">Rotor Speed (RPM)</Label>
                                <input type="text" class="form-control" id="modeRotSpeed"
                                    v-model.number="selectedPoint.RotSpeed" disabled>
                            </div>
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="modeWindSpeed" class="col-form-label">Wind Speed (m/s)</Label>
                                <input type="text" class="form-control" id="modeWindSpeed"
                                    v-model.number="selectedPoint.WindSpeed" disabled>
                            </div>
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="modeOP" class="col-form-label">Natural Frequency (Hz)</label>
                                <input type="text" class="form-control" id="modeOP"
                                    v-model.number="selectedPoint.NaturalFreqHz" disabled>
                            </div>
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="modeID" class="col-form-label">Damped Frequency (Hz)</Label>
                                <input type="text" class="form-control" id="modeID"
                                    v-model.number="selectedPoint.DampedFreqHz" disabled>
                            </div>
                            <div class="col-12 col-md-6 col-xl-4">
                                <label for="modeRotSpeed" class="col-form-label">Damping Ratio (%)</Label>
                                <input type="text" class="form-control" id="modeRotSpeed"
                                    v-model.number="selectedPoint.DampingRatio" disabled>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>

        <div class="card mb-3" v-if="selectedPoint != null && project.modeViz.length > 0">
            <div class="card-header hstack">
                <span>Mode Visualization</span>
                <a class="btn btn-primary btn-sm ms-auto" @click="project.clearModeViz">
                    Clear
                </a>
            </div>
            <div class="card-body">
                <div class="row">
                    <div class="col-9">
                        <div style="width:100%; height: 500px">
                            <ModeViz :key="clickedMVIdx" :ModeData="project.modeViz[clickedMVIdx]">
                            </ModeViz>
                        </div>
                    </div>
                    <div class="col-3">
                        <div class="list-group">
                            <a class="list-group-item list-group-item-action" v-for="mv in project.modeViz"
                                @click="selectModeViz(mv.OPID, mv.ModeID)">
                                OP: {{ mv.OPID }}, Mode: {{ mv.ModeID }}
                            </a>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </main>
</template>

<style scoped></style>
