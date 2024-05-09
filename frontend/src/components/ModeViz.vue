<script lang="ts" setup>
import { reactive, ref, onMounted, computed, watch } from 'vue'
import { main, viz } from '../../wailsjs/go/models';
import { useProjectStore } from '../project';
import * as THREE from 'three'

const project = useProjectStore()

const props = defineProps<{
    ModeData: viz.ModeData
}>()

watch(
    () => props.ModeData,
    (md) => {
        addFrames(md)
    },
    { deep: true }
)

let frames: THREE.Group[];
let frameNum = 0;
let frameCenter = new THREE.Vector3;
let frameSize = new THREE.Vector3;
let clock = new THREE.Clock();
let delta = 0;
const FOV = 10;

function addFrames(modeData: viz.ModeData) {
    console.log('ModeData in the canvas: ', modeData)
    scene.clear()
    const axesHelper = new THREE.AxesHelper(10);
    scene.add(axesHelper)
    // const geometry = new THREE.SphereGeometry(2, 32, 16);
    // const material = new THREE.MeshBasicMaterial({ color: 0xffff00 });
    // const sphere = new THREE.Mesh(geometry, material);
    // scene.add(sphere);
    frames = [] as THREE.Group[];
    const allFramesGroup = new THREE.Group()
    for (const f of modeData.Frames) {
        const frameGroup = new THREE.Group()
        for (const c of Object.values(f.Components)) {
            const curve = new THREE.CatmullRomCurve3(
                c.Line.map((p) => new THREE.Vector3(p.XYZ[0], p.XYZ[1], p.XYZ[2])))
            const points = curve.getPoints(50);
            const geometry = new THREE.BufferGeometry().setFromPoints(points);
            const material = new THREE.LineBasicMaterial({ color: 0xffffff, linewidth: 100 });
            const curveObject = new THREE.Line(geometry, material);
            frameGroup.add(curveObject)
            allFramesGroup.add(curveObject.clone()) // Add clone of object to be used for view sizing
        }
        frameGroup.visible = false // Initialize each group to not visible for animation
        frames.push(frameGroup)
        scene.add(frameGroup)
    }
    // scene.add(allFramesGroup)
    const bb = new THREE.Box3().setFromObject(allFramesGroup);
    frameCenter = bb.getCenter(new THREE.Vector3())
    frameSize = bb.getSize(new THREE.Vector3())
    frameNum = 0
}

let scene: THREE.Scene;
let renderer: THREE.WebGLRenderer;

const views = [
    {
        // Front View
        left: 0,
        bottom: 0,
        width: 0.495,
        height: 1.0,
        background: new THREE.Color().setRGB(0.3, 0.3, 0.3, THREE.SRGBColorSpace),
        eye: [-175, 0, frameCenter.z],
        up: [0, 0, 1],
        updateCamera: function (camera: THREE.PerspectiveCamera) {
            // Calculate distance along -X axis to fit model in frame vertically
            // See https://wejn.org/2020/12/cracking-the-threejs-object-fitting-nut/ for equation
            let distance = 1.05 * (frameSize.z / 2 / Math.tan(camera.fov * Math.PI / 180 / 2) + frameSize.x / 2)
            camera.position.fromArray([-distance, 0, frameCenter.z]); // Looking along X (downwind)
            camera.lookAt(frameCenter);
        },
        camera: new THREE.PerspectiveCamera,
    },
    {
        // Side View
        left: 0.5,
        bottom: 0,
        width: 0.245,
        height: 1.0,
        background: new THREE.Color().setRGB(0.3, 0.3, 0.3, THREE.SRGBColorSpace),
        eye: [0, -175, frameCenter.z],
        up: [0, 0, 1],
        updateCamera: function (camera: THREE.PerspectiveCamera) {
            // Calculate distance along -Y axis to fit model in frame vertically
            let distance = 1.05 * (frameSize.z / 2 / Math.tan(camera.fov * Math.PI / 180 / 2) + frameSize.y / 2)
            camera.position.fromArray([0, -distance, frameCenter.z]); // Looking along -Y (side)
            camera.lookAt(frameCenter);
        },
        camera: new THREE.PerspectiveCamera,
    },
    {
        // Top View
        left: 0.75,
        bottom: 0,
        width: 0.25,
        height: 1.0,
        background: new THREE.Color().setRGB(0.3, 0.3, 0.3, THREE.SRGBColorSpace),
        eye: [0, 0, 200],
        up: [0, 1, 0],
        updateCamera: function (camera: THREE.PerspectiveCamera) {
            // Calculate distance along Z axis to fit model in frame vertically
            let distance = 1.05 * (frameSize.y / 2 / Math.tan(camera.fov * Math.PI / 180 / 2) + frameSize.z)
            camera.position.fromArray([0, 0, distance]); // Looking along -Z (downward)
            camera.lookAt(frameCenter);
        },
        camera: new THREE.PerspectiveCamera,
    }
];

function animate() {
    render();
    delta += clock.getDelta()
    if (delta > 0.1) {
        delta = 0
        frames[frameNum].visible = false;
        frameNum++
        if (frameNum >= frames.length) frameNum = 0
        frames[frameNum].visible = true;
    }
    requestAnimationFrame(animate);
}

function render() {

    const canvas = renderer.domElement;
    const canvasWidth = canvas.clientWidth;
    const canvasHeight = canvas.clientHeight;
    if (canvas.width !== canvasWidth || canvas.height !== canvasHeight) {
        renderer.setSize(canvasWidth, canvasHeight, false);
    }

    for (let ii = 0; ii < views.length; ++ii) {

        const view = views[ii];
        const camera = view.camera;

        view.updateCamera(camera);

        const left = Math.floor(canvasWidth * view.left);
        const bottom = Math.floor(canvasHeight * view.bottom);
        const width = Math.floor(canvasWidth * view.width);
        const height = Math.floor(canvasHeight * view.height);

        renderer.setViewport(left, bottom, width, height);
        renderer.setScissor(left, bottom, width, height);
        renderer.setScissorTest(true);
        renderer.setClearColor(view.background);

        camera.aspect = width / height;
        camera.updateProjectionMatrix();

        renderer.render(scene, camera);
    }
}



onMounted(() => {

    const canvas = document.getElementById('modeVizCanvas')!;

    for (let ii = 0; ii < views.length; ++ii) {
        const view = views[ii];
        const camera = new THREE.PerspectiveCamera(FOV, 2, 1, 10000);
        camera.position.fromArray(view.eye);
        camera.up.fromArray(view.up);
        view.camera = camera;
    }

    scene = new THREE.Scene();
    addFrames(props.ModeData)

    renderer = new THREE.WebGLRenderer({ antialias: true, canvas });

    // document.addEventListener('mousemove', onDocumentMouseMove);

    animate();
})

</script>

<template>
    <canvas id="modeVizCanvas" class="h-100 w-100"></canvas>
</template>

<style scoped></style>
