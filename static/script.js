// JS File
'use strict';

const App = () => {
    // Store state for the server list
    const serverListState = van.state([])

    // TODO: Store this in local session or URL or something.
    const selectedServerState = van.state(null)

    // This loads the current configured servers and their running state, then it updates
    // the serverListState to update the rendered list.
    const loadServers = (async () => {
        const configs = await (await fetch('/api/config')).json()
        const runners = await (await fetch('/api/runner')).json()

        const tmpServerListState = []

        for (const serverName in configs.servers) {
            tmpServerListState.push({ name: serverName, running: (serverName in runners) })
        }

        serverListState.val = tmpServerListState
    })

    // Actually load the state
    loadServers()

    // And refresh the state every second
    setInterval(loadServers, 1000)

    // Update serverListState for the object with the name of `name` to running state of `state`
    const setServerListStateFor = (name, state) => {
        const tmpServerListState = []

        for (const item of serverListState.val) {
            if (item.name === name) {
                item.running = state
            }

            tmpServerListState.push(item)
        }

        serverListState.val = tmpServerListState
    }

    // Get the running state of the server with the name of `name`
    const getServerListStateFor = (name) => {
        for (const item of serverListState.val) {
            if (item.name === name) {
                return item.running
            }
        }

        return false
    }

    // Object to render the actual items in the server list
    const ServerItem = (name) => {
        const toggleServer = async () => {
            if (getServerListStateFor(name)) {
                await fetch(`/api/runner/${name}`, { method: 'DELETE' })
                setServerListStateFor(name, false)
                return
            }

            await fetch(`/api/runner/${name}`, { method: 'POST' })
            setServerListStateFor(name, true)
        }

        return van.tags.li(
            {
                class: () => selectedServerState.val === name ? 'server-item selected' : 'server-item',
                onclick: () => { selectedServerState.val = name }
            },
            name,
            van.tags.label(
                { class: 'switch', for: 'toggle-' + name },
                van.tags.input({
                    type: 'checkbox',
                    id: 'toggle-' + name,
                    checked: getServerListStateFor(name),
                    onclick: () => toggleServer(),
                }),
                van.tags.div({ class: 'slider' })
            )
        )
    }

    // Derive the server list state into a list of items to render
    const serverList = van.derive(() => van.tags.ul(
        { class: 'server-list' },
        serverListState.val.map((item) => ServerItem(item.name))
    ))

    // Derive the selection state to render the main viewer
    const mainViewer = van.derive(() => {
        const contentDom = (selectedServerState.val === null) ?
            van.tags.div({ id: 'frontpage' }, 'Select a server to view its logs :)') :
            'main viewer component for ' + selectedServerState.val

        return van.tags.div(
            { id: 'content' },
            contentDom
        )
    })

    return van.tags.div(
        { id: 'wrapper' },
        van.tags.nav(
            { id: 'nav' },
            van.tags.h1(
                { onclick: () => { selectedServerState.val = null } },
                'goprocmgr'
            ),
            serverList
        ),
        mainViewer,
    )
}

van.add(document.getElementById('app'), App());
