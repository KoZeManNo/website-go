// Regex to match message with one color
// /^ $/: Match start to end
// [0-8]: A number from 0 to 8
// [a-f0-9]{6}: 6 lowercase hex characters
const one = /^[0-8][a-f0-9]{6}$/
// Regex to match message with all colors
// /^ $/: Match start to end
// [a-f0-9]{54}: 6*9 lowercase hex characters
const all = /^[a-f0-9]{54}$/

const background = document.getElementById("background")
// Filtering background.childNodes divs
// because there are unwanted #text nodes
// (probably the newlines used for readability)
const panels = Array.from(background.childNodes).filter(node => {
    return node.nodeName === "DIV"
})

let socket



function updateColor(index, color) {
    color = "#" + color
    let panel = panels[index]
    let colorPicker = panel.firstChild

    panel.style.backgroundColor = color
    colorPicker.value = color
}

function connect() {
    // Get the protocol & URL required for a WebSocket connection
    let protocol = window.location.protocol === "https:" ? "wss://" : "ws://"
    let url = protocol + window.location.host + "/ws"

    socket = new WebSocket(url)
    
    socket.addEventListener("message", (event) => {
        let message = event.data
        
        if (all.test(message)) {
            // Split message every 6th character
            let colors = message.match(/.{1,6}/g)
            for (let i = 0; i < 9; i++) {
                updateColor(i, colors[i])
            }
        } else if (one.test(message)) {
            let index = message.charAt(0)
            let color = message.substring(1)
            updateColor(index, color)
        }
    })
}

// Reconnect socket when the window gets focused,
// because it often disconnects when it's not
window.onfocus = () => {
    if (socket.readyState === WebSocket.CLOSED) {
        connect()
    }
}

connect()



// For every panel (color) in the background,
// make the user click the color picker instead of the div
for (let i = 0; i < panels.length; i++) {
    let panel = panels[i]
    let colorPicker = panel.firstChild

    // Focus and click the color picker instead of the panel
    panel.addEventListener("click", () => {
        colorPicker.focus()
        colorPicker.click()
    })

    colorPicker.addEventListener("change", () => {
        socket.send(i + colorPicker.value.substring(1))
    })
}