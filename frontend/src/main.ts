import { mount } from 'svelte'
import App from './App.svelte'
import './less/style.less'

mount(App, { target: document.getElementById('app')! })
