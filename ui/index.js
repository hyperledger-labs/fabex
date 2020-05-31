let currentBlock;
let action;
let search;

let treeData = {
    name: "Blocks",
    children: []
}
let errorModal,
    input;

window.onload = function() {

    // define the tree-item component
    Vue.component("tree-item", {
        template: "#item-template",
        props: {
            item: Object
        },
        data: function() {
            return {
                isOpen: true
            };
        },
        computed: {
            isFolder: function() {
                return this.item.children && this.item.children.length;
            }
        },
        methods: {
            toggle: function() {
                if (this.isFolder) {
                    this.isOpen = !this.isOpen;
                }
            },
            makeFolder: function() {
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
            makeFolder: function(item) {
                let folder = Vue.set(item, "children", []);
                return folder
            },
            addItem: function(item, name) {
                item.children.push({
                    name: name
                });

                // возвращает внесённый элемент
                return item
            }
        }
    });

    let button = new Vue({
        el: '#button',
        methods: {
            getAns: async function() {

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
                    case "WriteSet key":
                        input.placeholder = 'payload key name'
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
    let min = 200000000;
    let max = -200000000;

    for (let i = 0; i < currentBlock.data.msg.length; i++) {
        if (currentBlock.data.msg[i].blocknum > max) max = currentBlock.data.msg[i].blocknum;
        if (currentBlock.data.msg[i].blocknum < min) min = currentBlock.data.msg[i].blocknum
    }

    let newBlockNum;

    if (param == 'left') newBlockNum = min - 1;
    if (param == 'right') newBlockNum = max + 1;

    action = 'Block number';
    search = newBlockNum;

    input.message = newBlockNum;

    GetBlockAndMakeTree();
}

async function GetBlockAndMakeTree() {

    treeData.children = [];

    try {
        if (action == "Block number") {
            var block = await axios.get(`http://localhost:5252/byblocknum/${search}`);
        } else if (action == "Tx ID") {
            var block = await axios.get(`http://localhost:5252/bytxid/${search}`);
        } else if (action == "WriteSet key") {
            var block = await axios.get(`http://localhost:5252/bykey/${search}`);
        }
    } catch (e) {
        errorModal.showModal = true;
        errorModal.httpCode = e;
        errorModal.error = e.response.data.error;
        return
    }

    currentBlock = block;

    block = block.data.msg;

    for (let i = 0; i < block.length; i++) {
        treeData.children.push({ name: "Block " + block[i].blocknum, children: [] });
        let folder = treeData.children[i].children;
        folder.push({ name: "channelid: " + block[i].channelid });
        folder.push({ name: "blockhash: " + block[i].blockhash });
        folder.push({ name: "previoushash: " + block[i].previoushash });
        folder.push({ name: "blocknum: " + block[i].blocknum });

        folder.push({ name: "txs", children: [] });
        folder = folder[4].children;

        for (let j = 0; j < block[i].txs.length; j++) {
            folder.push({ name: j, children: [] });
            let element = folder[j].children;
            element.push({ name: "txid: " + block[i].txs[j].txid });
            element.push({ name: "validationcode: " + block[i].txs[j].validationcode });
            element.push({ name: "time: " + block[i].txs[j].time });

            element.push({ name: "KV", children: [] });
            for (let x = 0; x < block[i].txs[j].KV.length; x++) {

                let kvFolder = element[3].children;
                kvFolder.push({ name: "key: " + block[i].txs[j].KV[x].key, children: [] });
                kvFolder = kvFolder[x].children;

                kvFolder.push({ name: "value: " + block[i].txs[j].KV[x].value });
            }
        }
    }
}