import { createApp } from 'vue'
import './style.css'
import App from './App.vue'

import 'vuefinder/dist/style.css';
import VueFinder from 'vuefinder';

import { VueQueryPlugin } from '@tanstack/vue-query'
import '@picocss/pico/css/pico.classless.blue.css'



const app = createApp(App)

// By default, Vuefinder will use English as the main language. 
// However, if you want to support multiple languages and customize the localization, 
// you can import the language files manually during component registration.
app.use(VueFinder)

app.use(VueQueryPlugin)

app.mount('#app')
 