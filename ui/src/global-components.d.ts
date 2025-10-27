// src/global-components.d.ts
import { VueFinder } from 'vuefinder'

declare module 'vue' {
  export interface GlobalComponents {
    VueFinder: typeof VueFinder,
  }
}