window.Store = new Vuex.Store({
  state: {
    connected: false,
    state: {}
  },
  mutations: {
    setConnected: (state, value) => {
      state.connected = value
    },
    setState: (state, value) => {
      state.state = value
    }
  }
})

window.Store.setConnecting = true

const wsURL = (window.location.protocol === 'https' ? 'wss://' : 'ws://') +
  window.location.host + '/ws'
window.Socket = new ReconnectingWebSocket(wsURL, null, {
  debug: true,
  reconnectInterval: 3000
})

window.Socket.addEventListener('open', () => {
  window.Store.commit('setConnected', true)
  console.log('WS connection opened')
})
window.Socket.addEventListener('message', (event) => {
  window.Store.commit('setState', JSON.parse(event.data))
})
window.Socket.addEventListener('close', () => {
  window.Store.commit('setConnected', false)
  console.log('WS connection closed')
})
