<script lang="ts" setup>
import { useProjectStore } from '../project';
import { reactive, onMounted } from 'vue'

const project = useProjectStore()

onMounted(() => { })

</script>

<template>
    <nav class="navbar navbar-expand-sm navbar-dark bg-primary fixed-top">
        <div class="container-fluid">
            <div class="collapse navbar-collapse" id="navbarNav">
                <ul class="navbar-nav me-auto">
                    <li class="nav-item">
                        <router-link class="nav-link" to="/">Project</router-link>
                    </li>
                    <li class="nav-item" v-if="project.loaded">
                        <router-link class="nav-link" to="/turbine">Turbine</router-link>
                    </li>
                    <li class="nav-item" v-if="project.loaded">
                        <router-link class="nav-link" to="/analysis">Analysis</router-link>
                    </li>
                    <li class="nav-item" v-if="project.loaded">
                        <router-link class="nav-link" to="/evaluate">Evaluate</router-link>
                    </li>
                    <li class="nav-item" v-if="project.loaded">
                        <router-link class="nav-link" to="/Results">Results</router-link>
                    </li>
                    <li class="nav-item" v-if="project.loaded">
                        <router-link class="nav-link" to="/vtk">VTK</router-link>
                    </li>
                </ul>
                <span class="navbar-text me-2" v-if="project.saving">
                    <div class="spinner-border spinner-border-sm text-light" role="status">
                        <span class="visually-hidden">Saving...</span>
                    </div>
                </span>
            </div>
        </div>
    </nav>
</template>

<style scoped></style>
