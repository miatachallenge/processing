const st = window.Store.state
const upstreamTypes = [
  'legacy_couchdb'
]

Vue.component('dashboard', {
  template: `<div class="container">
    <h1>
      Processing
      <div class="float-right">
        <button
          v-if="autostartParam !== 'true'"
          class="btn btn-md btn-success"
          @click="setAutostart('true')"
        >
          Enable autostart
        </button>
        <button
          v-else
          class="btn btn-md btn-danger"
          @click="setAutostart('false')"
        >
          Disable autostart
        </button>

        <button
          v-if="processing"
          class="btn btn-md btn-danger"
          @click="setProcessing(false)"
        >
          Stop processing
        </button>
        <button
          v-else
          class="btn btn-md btn-success"
          @click="setProcessing(true)"
        >
          Start processing
        </button>
      </div>
    </h1>
    <div class="row">
      <div class="col-lg-6">
        <h2>Source parameters</h2>
        <table class="table">
          <tr>
            <td>Local record count</td>
            <td>
              <code class="text-muted">{{ localCount }}</code>
            </td>
          </tr>
          <tr>
            <td style="width: 50%">
              <span style="line-height: 38px">
                Tag to process
              </span>
            </td>
            <td style="width: 50%">
              <code v-if="!tagEditing" @click="editTag">{{ tagParam || 'undefined' }}</code>
              <div v-else class="input-group">
                <select class="form-control" v-model="tagValue">
                  <option v-for="tag in availableTags" :value="tag.id">{{ tag.id }} - {{ tag.name }}</option>
                </select>
                <div class="input-group-append">
                  <button @click="saveTag" class="btn btn-outline-secondary">Save</button>
                </div>
              </div>
            </td>
          </tr>
          <tr>
            <td colspan="2">
              <h6><code>convert</code> function (arg is tag ID, return column value)</h6>
              <pre v-if="!convertEditing" @click="editConvert">{{ convertCode || 'undefined' }}</pre>
              <div v-else class="input-group">
                <textarea class="form-control" v-model="convertValue"></textarea>
                <div class="input-group-append">
                  <button @click="saveConvert" class="btn btn-outline-secondary">Save</button>
                </div>
              </div>
            </td>
          </tr>
          <tr>
            <td colspan="2">
              <h6><code>validate</code> function (validate milliseconds from arg, return bool)</h6>
              <pre v-if="!validateEditing" @click="editValidate">{{ validateCode || 'undefined' }}</pre>
              <div v-else class="input-group">
                <textarea class="form-control" v-model="validateValue"></textarea>
                <div class="input-group-append">
                  <button @click="saveValidate" class="btn btn-outline-secondary">Save</button>
                </div>
              </div>
            </td>
          </tr>
        </table>
      </div>
      <div class="col-lg-6">
        <h2>Synchronization target</h2>
        <table class="table">
          <tr>
            <td>Laps generated</td>
            <td>
              <code class="text-muted">{{ upstreamCount }}</code>
            </td>
          </tr>
          <tr>
            <td style="width: 50%">
              <span style="line-height: 38px">
                Type
              </span>
            </td>
            <td style="width: 50%">
              <code v-if="!upstreamEditing" @click="editUpstream">{{ upstreamParam || 'undefined' }}</code>
              <div v-else class="input-group">
                <select class="form-control" v-model="upstreamValue">
                  <option v-for="id in availableUpstreams" :value="id">{{ id }}</option>
                </select>
                <div class="input-group-append">
                  <button @click="saveUpstream" class="btn btn-outline-secondary">Save</button>
                </div>
              </div>
            </td>
          </tr>
          <tr>
            <td>
              <span style="line-height: 38px">
                Upstream address
              </span>
            </td>
            <td>
              <code v-if="!addressEditing" @click="editAddress">{{ addressParam || 'undefined' }}</code>
              <div v-else class="input-group">
                <input class="form-control" v-model="addressValue">
                <div class="input-group-append">
                  <button @click="saveAddress" class="btn btn-outline-secondary">Save</button>
                </div>
              </div>
            </td>
          </tr>
          <tr>
            <td>
              <span style="line-height: 38px">
                Drivers spreadsheet
              </span>
            </td>
            <td>
              <code v-if="!spreadsheetEditing" @click="editSpreadsheet">{{ spreadsheetParam || 'undefined' }}</code>
              <div v-else class="input-group">
                <input class="form-control" v-model="spreadsheetValue">
                <div class="input-group-append">
                  <button @click="saveSpreadsheet" class="btn btn-outline-secondary">Save</button>
                </div>
              </div>
            </td>
          </tr>
          <tr>
            <td>Drivers processed</td>
            <td>
              <code class="text-muted">{{ driversCount }}</code>
            </td>
          </tr>
        </table>
      </div>
    </div>
    <div class="row">
      <div class="col-lg-12">
        <h2>Most recent laps</h2>
        <table class="table">
          <thead>
            <tr>
              <th>Driver</th>
              <th>Completed</th>
              <th>Lap time</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in recentLaps">
              <td>{{ item.driver }}</td>
              <td>{{ item.timestamp | format }}</td>
              <td>{{ item.lap_time / 1000 }}s</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>`,
  data: () => {
    return {
      convertEditing: false,
      convertValue: null,
      validateEditing: false,
      validateValue: null,
      tagEditing: false,
      tagValue: null,
      spreadsheetEditing: false,
      spreadsheetValue: null,
      addressEditing: false,
      addressValue: null,
      upstreamEditing: false,
      upstreamValue: null,
      availableUpstreams: upstreamTypes
    }
  },
  methods: {
    editConvert () {
      this.convertValue = this.convertCode
      this.convertEditing = true
    },
    saveConvert () {
      window.Socket.send(JSON.stringify({
        'type': 'update',
        'param': 'convert',
        'value': this.convertValue
      }))
      this.convertEditing = false
    },
    editValidate () {
      this.validateValue = this.validateCode
      this.validateEditing = true
    },
    saveValidate () {
      window.Socket.send(JSON.stringify({
        'type': 'update',
        'param': 'validate',
        'value': this.validateValue
      }))
      this.validateEditing = false
    },
    editTag () {
      this.tagValue = this.tagParam
      this.tagEditing = true
    },
    saveTag () {
      window.Socket.send(JSON.stringify({
        'type': 'update',
        'param': 'tag',
        'value': this.tagValue
      }))
      this.tagEditing = false
    },
    editSpreadsheet () {
      this.spreadsheetValue = this.spreadsheetParam
      this.spreadsheetEditing = true
    },
    saveSpreadsheet () {
      window.Socket.send(JSON.stringify({
        'type': 'update',
        'param': 'legacy_spreadsheet',
        'value': this.spreadsheetValue
      }))
      this.spreadsheetEditing = false
    },
    editAddress () {
      this.addressValue = this.addressParam
      this.addressEditing = true
    },
    saveAddress () {
      window.Socket.send(JSON.stringify({
        'type': 'update',
        'param': 'legacy_address',
        'value': this.addressValue
      }))
      this.addressEditing = false
    },
    editUpstream () {
      this.upstreamValue = this.upstreamParam
      this.upstreamEditing = true
    },
    saveUpstream () {
      window.Socket.send(JSON.stringify({
        'type': 'update',
        'param': 'upstream',
        'value': this.upstreamValue
      }))
      this.upstreamEditing = false
    },
    setAutostart (newValue) {
      window.Socket.send(JSON.stringify({
        'type': 'update',
        'param': 'autostart',
        'value': newValue
      }))
    },
    setProcessing (newValue) {
      window.Socket.send(JSON.stringify({
        'type': 'processing',
        'value': newValue ? 'true' : 'false'
      }))
    }
  },
  computed: {
    processing () {
      return st.state && st.state.upstream && st.state.upstream.processing
    },
    autostartParam () {
      return st.state && st.state.params && st.state.params.autostart
    },
    localCount () {
      return st.state && st.state.database && st.state.database.count
    },
    upstreamCount () {
      return st.state && st.state.upstream && st.state.upstream.count
    },
    tagParam () {
      return st.state && st.state.params && st.state.params.tag
    },
    availableTags () {
      return (st.state && st.state.tags) || []
    },
    convertCode () {
      return st.state && st.state.params && st.state.params.convert
    },
    validateCode () {
      return st.state && st.state.params && st.state.params.validate
    },
    addressParam () {
      return st.state && st.state.params && st.state.params.legacy_address
    },
    spreadsheetParam () {
      return st.state && st.state.params && st.state.params.legacy_spreadsheet
    },
    upstreamParam () {
      return st.state && st.state.params && st.state.params.upstream
    },
    driversCount () {
      return st.state && st.state.legacy_spreadsheet && st.state.legacy_spreadsheet.count
    },
    recentLaps () {
      return st.state && st.state.recent_laps && st.state.recent_laps.length > 0 && st.state.recent_laps.reverse() || []
    }
  },
  filters: {
    format(value) {
      return moment(value).format("DD/MM/YY HH:mm:ss")
    }
  }
})