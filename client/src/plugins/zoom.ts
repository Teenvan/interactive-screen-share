import { PluginObject } from 'vue'
import { NekoClient } from '~/neko'

declare global {
  const $zoomSdk: NekoClient

  interface Window {
    $zoomSdk: NekoClient
  }
}

declare module 'vue/types/vue' {
  interface Vue {
    $zoomSdk: NekoClient
  }
}

const plugin: PluginObject<undefined> = {
  install(Vue) {
    window.$zoomSdk = new NekoClient()
      .on('error', window.$log.error)
      .on('warn', window.$log.warn)
      .on('info', window.$log.info)
      .on('debug', window.$log.debug)

    Vue.prototype.$zoomSdk = window.$zoomSdk
  },
}

export default plugin
