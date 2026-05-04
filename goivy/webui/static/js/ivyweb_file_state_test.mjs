import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';

const appSource = fs.readFileSync(new URL('./ivyweb_app.js', import.meta.url), 'utf8');

const sandbox = {
    console,
    window: {},
    document: {
        getElementById() {
            return null;
        },
        addEventListener() {},
    },
    IvyAPI: class {
        constructor() {
            this.sessionId = 's1';
        }
    },
    IvyControls: class {
        setStatus(message, kind) {
            this.lastStatus = { message, kind };
        }
    },
    IvyGraph: class {},
    IvyPersist: {
        getSessionIdFromURL() {
            return 's1';
        },
        async loadFileHandle() {
            throw new Error('test must override loadFileHandle');
        },
        async saveFileHandle() {},
        save() {},
        setFileName() {},
    },
};

vm.createContext(sandbox);
vm.runInContext(appSource + '\nthis.IvyApp = IvyApp;', sandbox);

async function testSaveRecoversPersistedHandleBeforeSaveAs() {
    let loadFileHandleCalled = false;
    let saveAsCalled = false;
    let written = '';
    const fakeHandle = {
        name: 'helloworld.ivy',
        async createWritable() {
            return {
                async write(content) {
                    written = content;
                },
                async close() {},
            };
        },
    };

    sandbox.IvyPersist.loadFileHandle = async function (state) {
        loadFileHandleCalled = true;
        assert.equal(state.sessionId, 's1');
        assert.equal(state.fileName, 'helloworld.ivy');
        assert.equal(state.filePath, 'helloworld.ivy');
        return fakeHandle;
    };

    const app = new sandbox.IvyApp();
    app._persistedFileName = 'helloworld.ivy';
    app._persistedFilePath = 'helloworld.ivy';
    app._persistedFileContent = 'old content';
    app._savedFileContent = 'old content';
    app.cmEditor = { getValue: () => 'old content ' };
    app._updateEditorLabel = function () {};
    app._ensureFileHandleWritable = async function () { return true; };
    app._confirmNoExternalChangeBeforeSave = async function () { return 'ok'; };
    app.saveAs = async function () {
        saveAsCalled = true;
        return false;
    };

    const saved = await app.save();

    assert.equal(saved, true);
    assert.equal(loadFileHandleCalled, true);
    assert.equal(saveAsCalled, false);
    assert.equal(app._fileHandle, fakeHandle);
    assert.equal(written, 'old content ');
}

await testSaveRecoversPersistedHandleBeforeSaveAs();
