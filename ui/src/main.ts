import { createApp } from 'vue'
import './style.css'
import App from './App.vue'

import VueFinder from 'vuefinder/dist/vuefinder'
import 'vuefinder/dist/style.css'

const app = createApp(App)

// By default, Vuefinder will use English as the main language. 
// However, if you want to support multiple languages and customize the localization, 
// you can import the language files manually during component registration.
app.use(VueFinder)

app.mount('#app')
 