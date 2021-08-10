import {id} from './manifest';

class Plugin {
    initialize(registry, store) {}
}

window.registerPlugin(id, new Plugin());
