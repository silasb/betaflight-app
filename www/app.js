// import { h, app } from "hyperapp"

const h = window.hyperapp.h

let state = betaflight.data;

const actions = {
  incr: ([value, flightSurface, pid]) => {
    betaflight.incrPid(value, flightSurface, pid)
  },
  dec: ([value, flightSurface, pid]) => {
    betaflight.decPid(value, flightSurface, pid)
  },
  updateGlobalState: value => state => value
}

const view = (state, actions) => (
  <div>
    <h1>{JSON.stringify(state)}</h1>
    {Object.keys(state.flightSurfaces).map(key => {
      const flightSurface = state.flightSurfaces[key]

      return <div>
        <h1>{flightSurface.name}</h1>

        <div className="parent">
          {Object.keys(state.flightSurfaces[key].pids).map(pidKey => {
            const pid = state.flightSurfaces[key].pids[pidKey]

            return <div className="pid2">
              <div>{pid.name}</div>

              <div className="pid">
                <button  onclick={() => actions.incr([1, flightSurface.name.toLowerCase(), pid.name.toLowerCase()])}>+</button>
                {pid.value}
                <button onclick={() => actions.dec([1, flightSurface.name.toLowerCase(), pid.name.toLowerCase()])}>-</button>
              </div>
            </div>
          })}
        </div>
      </div>
    })}
  </div>
)

const app = window.hyperapp.app(state, actions, view, document.body)

betaflight.render = () => {
  app.updateGlobalState(betaflight.data)
}