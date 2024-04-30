<script lang="ts" setup>
import { useProjectStore } from '../project';
import { useWindowSize } from '@vueuse/core';
import { Line, AxesHelper, BoxGeometry, BufferGeometry, CubeCamera, DirectionalLight, LineBasicMaterial, Mesh, MeshBasicMaterial, MeshLambertMaterial, MeshStandardMaterial, Object3D, Object3DEventMap, PerspectiveCamera, RenderTarget, Scene, SphereGeometry, Vector3, WebGLRenderer, _SRGBAFormat, GridHelper, Color, PointsMaterial, Points, Float32BufferAttribute, Vector3Tuple, TubeGeometry, ShapeGeometry, Shape, Vector2 } from 'three';
import { computed, onMounted, ref, render, watch } from 'vue';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { VTKLoader } from 'three/examples/jsm/loaders/VTKLoader.js';
import { LogPrint, LogTrace } from '../wailsjs/runtime/runtime';
import { GetVTKFile, LoadVTKFile } from "../../wailsjs/go/main/App"
import { LogError } from '../../wailsjs/runtime/runtime';

const project = useProjectStore()

// Load and parse VTK files
// Convert type to vector
function convertToVertices (points: Float32Array[]) {
    const vertices = []
    for (let index = 0; index < points.length; index++) {
        vertices.push(
            new Vector3(
                points[index][0],
                points[index][1],
                points[index][2]
            )
        )
    }

    return vertices
}

// Draw Points
function drawPoints (pointsAsVectors: Vector3[]) {
    const geometry = new BufferGeometry().setFromPoints(pointsAsVectors)
    const material = new PointsMaterial( {color: 0xff00ff, size: 2} )
    const points = new Points(geometry, material)
    console.log('points: ', points)

    return points
}

// Draw Lines
function drawLines (pointsAsVectors: Vector3[], connectivity: Int32Array, offsets: Int32Array) {
    
    const pointPairsToDraw = []
    const material = new LineBasicMaterial({
        color: 0xff0000,
        linewidth: 5
    }); // TODO: control line thickness other way

    console.log(connectivity)
    console.log(offsets)

    pointPairsToDraw.push(pointsAsVectors[0])
    offsets.forEach(i => 
        pointPairsToDraw.push(pointsAsVectors[connectivity[i-1]])
    )
    
    const geometry = new BufferGeometry().setFromPoints(pointPairsToDraw)
    const line = new Line(geometry, material)
    console.log('line: ', line)

    return line
}

// Draw Polygons
function drawPolygons (pointsAsVectors: Vector3[], connectivity: Int32Array, offsets: Int32Array) {
    // TODO: ShapeGeometry only supports for Vector2
    const shape = new Shape()
    offsets.forEach(numCorners => {
        shape.moveTo(pointsAsVectors[connectivity[0]].x, pointsAsVectors[connectivity[0]].y)
        for (let index = 1; index < numCorners; index++) {
            shape.lineTo(pointsAsVectors[connectivity[index]].x, pointsAsVectors[connectivity[index]].y)
        }
    })

    const geometry = new ShapeGeometry(shape)
    const material = new MeshBasicMaterial( { color: 0x00ff00 } )
    const poly = new Mesh(geometry, material)
    console.log('polygon: ', poly)

    return poly
}

// Threejs
let renderer: WebGLRenderer
let camera: PerspectiveCamera
let controls: OrbitControls
const experience = ref<HTMLCanvasElement | null>(null)

// Scene with camera, light
const scene = new Scene()

// Load VTK -- import from THREE.js
// const loader = new VTKLoader()

// Add loop
const loop = () => {
    renderer.render(scene, camera)
    controls.update()
    // cube.position.y += 0.01     // every loop update the position
    requestAnimationFrame(loop)
}

onMounted(() => {
    const { width, height } = useWindowSize()
    const aspectRatio = computed(() => width.value / height.value)

    // Watch window resize
    function updateRenderer() {
        renderer.setSize(width.value, height.value)     // Full screen
        renderer.setPixelRatio(window.devicePixelRatio)
    }

    function updateCamera() {
        camera.aspect = aspectRatio.value
        camera.updateProjectionMatrix()

    }
    watch(aspectRatio, updateRenderer)
    watch(aspectRatio, updateCamera)
    
    camera = new PerspectiveCamera(
        60,
        aspectRatio.value,
        0.1,
        1000
    )
    camera.position.set(60, 60, 60)
    camera.lookAt(new Vector3(60,60,0))
    camera.up = new Vector3(0,0,1)
    scene.add(camera)

    const light = new DirectionalLight(0xffffff, 1)
    light.position.z = 5
    light.position.x = -2
    scene.add(light)

    // Draw axes and grids for visualizing dimensions in the scene
    const zxGridHelper = new GridHelper(300, 10)
    scene.add(zxGridHelper)

    const yzGridHelper = new GridHelper(300, 10)
    yzGridHelper.rotation.z = Math.PI/2
    scene.add(yzGridHelper)

    const xyGridHelper = new GridHelper(300, 10)
    xyGridHelper.rotation.x = Math.PI/2
    scene.add(xyGridHelper)

    const axesHelper = new AxesHelper(150)       // X:Red, Y:Green, Z:Blue
    scene.add(axesHelper)
    

    // Draw VTK Geometries
    LoadVTKFile().then(resultList => {
        console.log(resultList)
        resultList.forEach((result, index) => {
            const points = result.PolyData.Piece.Points.DataArray.MatrixF32
            const pointsAsVectors = convertToVertices(points)

            // Draw Points
            if (result.PolyData.Piece['NumberOfPoints'] != 0) {
                const pointToDraw = drawPoints(pointsAsVectors)
                scene.add(pointToDraw)
            }

            // Draw Lines
            if (result.PolyData.Piece['NumberOfLines'] != 0) {
                const connectivity = result.PolyData.Piece.Lines.DataArray[0].Connectivity
                const offsets = result.PolyData.Piece.Lines.DataArray[1].Offsets
                const line = drawLines(pointsAsVectors, connectivity, offsets)
                scene.add(line)
            }

            // Draw Polygons
            if (result.PolyData.Piece['NumberOfPolys'] != 0) {
                const connectivity = result.PolyData.Piece.Polys.DataArray[0].Connectivity
                const offsets = result.PolyData.Piece.Polys.DataArray[1].Offsets
                const polys = drawPolygons(pointsAsVectors, connectivity, offsets)
                scene.add(polys)
            }

        })
    }).catch(err => {
        LogError(err)
        console.log(err)
    })

    renderer = new WebGLRenderer({
        canvas: experience.value as unknown as HTMLCanvasElement,
        antialias: true,
        alpha: true
    })
    
    // Add controls - rotating, zooming, etc.
    controls = new OrbitControls(camera, renderer.domElement)
    controls.enableDamping = true
    // Adapt to the changing window size (shape might change accordingly)
    updateRenderer()
    updateCamera()
    // Shape is maintained as it's rendered another time
    loop()
})
</script>

<template>
    <canvas ref="experience"></canvas>
</template>








