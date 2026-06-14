import {CheckUpdate, ShowUpdatePrompt, GetConfig, SaveURL} from '../wailsjs/go/main/App';
import logo from './assets/images/logo-universal.png';

let currentConfig = null;

// CSS Styles
const style = document.createElement('style');
style.innerHTML = `
    body { background-color: #ffffff; margin: 0; padding: 0; }
    .spinner { border: 4px solid rgba(0, 0, 0, 0.05); width: 36px; height: 40px; border-radius: 50%; border-left-color: #00CCFF; animation: spin 1s linear infinite; margin: 0 auto 20px auto; }
    @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
    .manual-link { color: #007BFF; text-decoration: underline; cursor: pointer; font-size: 13px; opacity: 0.7; }
    .manual-link:hover { opacity: 1; }
    .btn-save { padding: 12px 24px; background: #00CCFF; color: white; border: none; border-radius: 8px; cursor: pointer; font-weight: bold; margin-top: 15px; width: 100%; font-size: 16px; }
    .btn-save:hover { background: #00b3db; }
    .input-url { padding: 12px; width: 100%; box-sizing: border-box; border: 1px solid #ddd; border-radius: 8px; margin-bottom: 15px; outline: none; font-size: 15px; }
    .input-url:focus { border-color: #00CCFF; }
    .card { background: white; padding: 30px; border-radius: 12px; box-shadow: 0 10px 25px rgba(0,0,0,0.05); width: 90%; max-width: 400px; }
`;
document.head.appendChild(style);

function showLoadingUI() {
    document.querySelector('#app').innerHTML = `
        <div style="display: flex; justify-content: center; align-items: center; height: 100vh; font-family: 'Segoe UI', sans-serif; flex-direction: column; text-align: center;">
            <img src="${logo}" style="width: 100px; margin-bottom: 40px;">
            <div class="spinner"></div>
            <h3 style="margin: 0 0 10px 0; font-weight: 500; color: #444;">Menghubungkan ke Vines POS</h3>
            <p style="color: #888; font-size: 14px; margin-bottom: 40px;">Mohon tunggu sebentar...</p>
            
            <div style="position: fixed; bottom: 30px; width: 100%;">
                <span class="manual-link" id="btn-settings">Ubah Alamat Server (URL)</span>
            </div>
        </div>
    `;
    const btnSettings = document.getElementById('btn-settings');
    if (btnSettings) btnSettings.onclick = showSettingsUI;
}

function showSettingsUI() {
    document.querySelector('#app').innerHTML = `
        <div style="display: flex; justify-content: center; align-items: center; height: 100vh; font-family: 'Segoe UI', sans-serif; background-color: #f8f9fa;">
            <div class="card">
                <img src="${logo}" style="width: 60px; margin-bottom: 20px;">
                <h2 style="margin: 0 0 10px 0; font-size: 22px;">Konfigurasi Server</h2>
                <p style="color: #666; font-size: 14px; margin-bottom: 25px;">Masukkan alamat URL server POS Anda untuk memulai koneksi.</p>
                
                <div style="text-align: left;">
                    <label style="font-size: 12px; font-weight: bold; color: #999; text-transform: uppercase; margin-bottom: 5px; display: block;">URL Server</label>
                    <input type="text" id="url-input" class="input-url" placeholder="Contoh: http://45.64.97.50:888/thevines" value="${currentConfig?.remote_url || ''}">
                </div>
                
                <button id="save-btn" class="btn-save">Simpan & Hubungkan</button>
                
                ${currentConfig?.remote_url ? `
                    <div style="margin-top: 20px;">
                        <span class="manual-link" onclick="window.location.reload()">Batal & Kembali</span>
                    </div>
                ` : ''}
            </div>
        </div>
    `;
    
    document.getElementById('save-btn').onclick = () => {
        let newURL = document.getElementById('url-input').value.trim();
        if (!newURL.startsWith('http')) {
            alert('URL tidak valid! Gunakan format http:// atau https://');
            return;
        }
        
        SaveURL(newURL).then(res => {
            if (res === "Success") {
                // Reload ke root agar muncul loading UI dulu sebelum redirect
                window.location.href = '/';
            } else {
                alert("Gagal menyimpan: " + res);
            }
        });
    };
}

// Ekspos ke window agar bisa dipanggil dari Go (Menu Bar)
window.showSettingsUI = showSettingsUI;
// Main Execution
showLoadingUI();

const urlParams = new URLSearchParams(window.location.search);
const forceReset = urlParams.get('reset') === '1';

GetConfig().then(config => {
    currentConfig = config;

    // 1. Cek Update di background
    CheckUpdate().then(result => {
        if (result.update_available) ShowUpdatePrompt(result.latest_version, result.release);
    });

    // 2. Logika Redirect
    if (!config.remote_url || forceReset) {
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
