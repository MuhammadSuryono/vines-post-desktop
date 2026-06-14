import {CheckUpdate, ShowUpdatePrompt, GetConfig, SaveURL} from '../wailsjs/go/main/App';
import logo from './assets/images/logo-universal.png';

let currentConfig = null;

// CSS Styles
const style = document.createElement('style');
style.innerHTML = `
    .spinner { border: 4px solid rgba(0, 0, 0, 0.1); width: 40px; height: 40px; border-radius: 50%; border-left-color: #00CCFF; animation: spin 1s linear infinite; margin: 0 auto 20px auto; }
    @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
    .manual-link { color: #007BFF; text-decoration: underline; cursor: pointer; font-size: 14px; }
    .btn-save { padding: 10px 20px; background: #00CCFF; color: white; border: none; border-radius: 5px; cursor: pointer; font-weight: bold; margin-top: 10px; }
    .input-url { padding: 10px; width: 80%; max-width: 400px; border: 1px solid #ccc; border-radius: 5px; margin-bottom: 10px; outline: none; }
`;
document.head.appendChild(style);

function showLoadingUI() {
    document.querySelector('#app').innerHTML = `
        <div style="display: flex; justify-content: center; align-items: center; height: 100vh; font-family: sans-serif; flex-direction: column; text-align: center;">
            <img src="${logo}" style="width: 120px; margin-bottom: 30px;">
            <div class="spinner"></div>
            <h2>Menghubungkan ke Vines POS...</h2>
            <div id="retry-area" style="margin-top: 30px; display: none;">
                <span class="manual-link" id="btn-settings">Ubah Pengaturan URL</span>
            </div>
        </div>
    `;
    document.getElementById('btn-settings').onclick = showSettingsUI;
}

function showSettingsUI() {
    document.querySelector('#app').innerHTML = `
        <div style="display: flex; justify-content: center; align-items: center; height: 100vh; font-family: sans-serif; flex-direction: column; text-align: center; padding: 20px;">
            <img src="${logo}" style="width: 80px; margin-bottom: 20px;">
            <h3>Pengaturan Server POS</h3>
            <p style="color: #666; font-size: 14px;">Masukkan alamat URL server (misal: http://1.2.3.4:888/path)</p>
            <input type="text" id="url-input" class="input-url" placeholder="http://..." value="${currentConfig?.remote_url || ''}">
            <button id="save-btn" class="btn-save">Simpan & Hubungkan</button>
            ${currentConfig?.remote_url ? '<p><span class="manual-link" onclick="location.reload()">Batal</span></p>' : ''}
        </div>
    `;
    document.getElementById('save-btn').onclick = () => {
        const newURL = document.getElementById('url-input').value;
        if (!newURL.startsWith('http')) {
            alert('URL harus diawali dengan http:// atau https://');
            return;
        }
        SaveURL(newURL).then(res => {
            if (res === "Success") {
                location.reload();
            } else {
                alert(res);
            }
        });
    };
}

// Main Execution
showLoadingUI();

GetConfig().then(config => {
    currentConfig = config;
    
    // 1. Cek Update di background
    CheckUpdate().then(result => {
        if (result.update_available) ShowUpdatePrompt(result.latest_version, result.url);
    });

    // 2. Logika Redirect
    if (!config.remote_url) {
        showSettingsUI();
    } else {
        setTimeout(() => {
            // Munculkan opsi settings jika loading kelamaan
            setTimeout(() => {
                const retry = document.getElementById('retry-area');
                if(retry) retry.style.display = 'block';
            }, 4000);
            window.location.assign(config.remote_url);
        }, 500);
    }
});
