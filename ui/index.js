let action;
let search;

let treeData = {
    name: "Blocks",
    children: []
}
let errorModal,
    input;

// Initialization of Vuew.js components
window.onload = function () {

    // define the tree-item component
    Vue.component("tree-item", {
        template: "#item-template",
        props: {
            item: Object
        },
        data: function () {
            return {
                isOpen: true
            };
        },
        computed: {
            isFolder: function () {
                return this.item.children && this.item.children.length;
            }
        },
        methods: {
            toggle: function () {
                if (this.isFolder) {
                    this.isOpen = !this.isOpen;
                }
            },
            makeFolder: function () {
                if (!this.isFolder) {
                    this.$emit("make-folder", this.item);
                    this.isOpen = true;
                }
            }
        }
    });

    // boot up the demo
    var demo = new Vue({
        el: "#demo",
        data: {
            treeData: treeData
        },
        methods: {
            makeFolder: function (item) {
                let folder = Vue.set(item, "children", []);
                return folder
            },
            addItem: function (item, name) {
                item.children.push({
                    name: name
                });

                // returns the inserted item
                return item
            }
        }
    });

    let button = new Vue({
        el: '#button',
        methods: {
            getAns: async function () {

                action = checkbox.block;
                search = input.message;

                if (search == false) {
                    return
                }

                GetBlockAndMakeTree();
            }
        }
    })

    let checkbox = new Vue({
        el: '#checkbox',
        data: {
            block: 'Block number'
        },
        methods: {
            onChange(event) {
                switch (event.target.value) {
                    case "Block number":
                        input.placeholder = '1'
                        break;
                    case "Tx ID":
                        input.placeholder = 'af589062d2e699c9b0ba36e831609876f0ebae99'
                        break;
                    default:
                        input.placeholder = '1'
                }
            }
        }
    });

    input = new Vue({
        el: '#input',
        data: {
            message: '',
            placeholder: '1'
        }
    });

    // register modal component
    Vue.component("modal", {
        template: "#modal-template",
    });

    // start app
    errorModal = new Vue({
        el: "#app",
        data: {
            showModal: false,
            httpCode: '',
            error: ''
        }
    });
};

function GetNewBlock(param) {
    var newBlockNum;

    if (param == 'left') newBlockNum = parseInt(input.message) - 1;
    if (param == 'right') newBlockNum = parseInt(input.message) + 1;

    if (newBlockNum < 0) newBlockNum = 0

    action = 'Block number';
    search = newBlockNum;

    input.message = newBlockNum;

    GetBlockAndMakeTree();
}

async function GetBlockAndMakeTree() {

    if (parseInt(search) < 0) {
        input.message = 0
        search = 0
    }

    treeData.children = [];

    var err = false;

    try {
        if (action == "Block number") {
            var ans = await axios.get(`http://localhost:5252/byblocknum/${search}`);
        } else if (action == "Tx ID") {
            var ans = await axios.get(`http://localhost:5252/bytxid/${search}`);
        }
    } catch (e) {
        errorModal.showModal = true;
        errorModal.httpCode = e;
        errorModal.error = e.response.data.error;
        err = true;
    }

    if (err) {
        return
    }

    block = ans.data.msg;

    console.log(block)

    // parsing json into a tree
    treeData.children.push({name: "Block " + block.blocknum, children: []});
    let folder = treeData.children[0].children;
    folder.push({name: "channelid: " + block.channelid});
    folder.push({name: "blockhash: " + block.blockhash});
    folder.push({name: "previoushash: " + block.previoushash});
    folder.push({name: "blocknum: " + block.blocknum});

    folder.push({name: "txs", children: []});
    folder = folder[4].children;

    for (let j = 0; j < block.txs.length; j++) {
        folder.push({name: j, children: []});
        let element = folder[j].children;
        element.push({name: "txid: " + block.txs[j].txid});
        element.push({name: "validationcode: " + block.txs[j].validationcode});
        element.push({name: "time: " + block.txs[j].time});

        element.push({name: "KV", children: []});

        var isConfig = false;

        for (let x = 0; x < block.txs[j].KV.length; x++) {
            var value = window.atob(block.txs[j].KV[x].value)
            if (block.txs[j].KV[x].key == "Type" && value == "Config") {
                isConfig = true;
                break;
            }
        }

        for (let x = 0; x < block.txs[j].KV.length; x++) {
            let kvFolder = element[3].children;
            kvFolder.push({name: "key: " + block.txs[j].KV[x].key, children: []});
            kvFolder = kvFolder[x].children;

            var value = window.atob(block.txs[j].KV[x].value)

            if (isConfig) {
                var key = block.txs[j].KV[x].key
                if (key == "Groups" || key == "Values" || key == "Policies") {
                    value = JSON.parse(value);
                    GoDeep(value)
                    value = JSON.stringify(value)
                }
            }

            kvFolder.push({name: "value: " + value});
        }
    }
}

function HasNesting(jsonData) {
    if (typeof jsonData == "object") {
        return true
    }

    return false
}

function GoDeep(jsonData) {
    Object.keys(jsonData).forEach(function (key) {
        var value = jsonData[key];

        if (HasNesting(value)) {
            GoDeep(value)
            return
        }

        if (key !== "value") {
            return
        }

        try {
            var dec = window.atob(value)
            jsonData[key] = dec
        } catch (e) {
            // do nothing
        }

    });
}