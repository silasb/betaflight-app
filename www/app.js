// import { h, app } from "hyperapp"

const h = window.hyperapp.h

let state = {
  data: betaflight.data,
}

const actions = {
  incr: ([value, flightSurface, pid]) => {
    betaflight.incrPid(value, flightSurface, pid)
  },
  dec: ([value, flightSurface, pid]) => {
    betaflight.decPid(value, flightSurface, pid)
  },
  savePids: value => (state, actions) => {
    betaflight.savePids()

    setTimeout(actions.data.set, 2000, '')
  },
  connect: value => (state, actions) => {
    betaflight.connect(value)

    setTimeout(actions.data.set, 2000, '')
  },
  data: {
    set: value => state => {
      betaflight.setFlash(value)

      return { flash: value }
    },
  },
  updateGlobalState: value => state => value,
  updateBetaflightData: value => state => ({data: value})
}

const rank = {
  'p': 2,
  'i': 1,
  'd': 0
}

const ConnectionView = ({serialPorts, connect}) => (
  <div className="serial-ports">
    <div className="banner">hi tune your pids with this app</div>
    <div className="banner">please select the serial port you want to connect to:</div>
      {serialPorts.map(m => (
        <div className="serial-port" onclick={() => connect(m)}>{m}</div>
      ))}
  </div>
)

const PidControlView = ({savePids, flightSurfaces, incr, dec}) => (
  <div>
    <button onclick={() => savePids([])}>Save</button>

    {Object.keys(flightSurfaces).map(key => {
      const flightSurface = flightSurfaces[key]

      return <div class="pids">
        <h1>{flightSurface.name}</h1>

        <div className="parent">
          {Object.keys(flightSurfaces[key].pids).sort((a, b) => (rank[b] - rank[a])).map(pidKey => {
            const pid = flightSurfaces[key].pids[pidKey]

            return <div className="pid2">
              <div>{pid.name}</div>

              <div className="pid">
                <button  onclick={() => incr([1, flightSurface.name.toLowerCase(), pid.name.toLowerCase()])}>+</button>
                {pid.value}
                <button onclick={() => dec([1, flightSurface.name.toLowerCase(), pid.name.toLowerCase()])}>-</button>
              </div>
            </div>
          })}
        </div>
      </div>
    })}
  </div>
)

const Flash = ({message}) => {
  if (message) {
    return <div className="flash">
      {message}
    </div>
  }
}

const view = (state, actions) => (
  <div>
    {/* <p className="debug">{JSON.stringify(state)}</p> */}

    <Flash message={state.data.flash} />

    {state.data.connectedSerialPort == ""
      ? <ConnectionView serialPorts={state.data.serialPortsAvailable} connect={actions.connect} saveSerial={actions.saveSerial} />
      : <PidControlView savePids={actions.savePids} flightSurfaces={state.data.flightSurfaces} incr={actions.incr} dec={actions.dec} />
    }

  </div>
)

const app = window.hyperapp.app(state, actions, view, document.body)

betaflight.render = () => {
  app.updateBetaflightData(betaflight.data)
}

const debug = (blah) => alert(JSON.stringify(blah))