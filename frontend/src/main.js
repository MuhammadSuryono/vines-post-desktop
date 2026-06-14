import {CheckUpdate, ShowUpdatePrompt} from '../wailsjs/go/main/App';
import logo from './assets/images/logo-universal.png';

const remoteURL = "http://45.64.97.50:888/thevines/index.php";

// Menambahkan style untuk spinner animasi langsung ke dalam halaman
const style = document.createElement('style');
style.innerHTML = `
    .spinner {
        border: 4px solid rgba(0, 0, 0, 0.1);
        width: 40px;
        height: 40px;
        border-radius: 50%;
        border-left-color: #00CCFF;
        animation: spin 1s linear infinite;
        margin: 0 auto 20px auto;
    }
    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
    }
    .manual-link {
        color: #007BFF;
        text-decoration: underline;
        cursor: pointer;
        font-size: 14px;
    }
    .manual-link:hover {
        color: #0056b3;
    }
`;
document.head.appendChild(style);

document.querySelector('#app').innerHTML = `
    <div style="display: flex; justify-content: center; align-items: center; height: 100vh; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; flex-direction: column; text-align: center; background-color: #ffffff; color: #333;">
        <img src="${logo}" alt="Logo" style="width: 120px; margin-bottom: 30px;">
        <div class="spinner"></div>
        <h2 id="status" style="margin: 0 0 10px 0; font-weight: 500;">Menghubungkan ke Vines POS...</h2>
        <p id="sub-status" style="color: #666; margin: 0;">Memuat antarmuka sistem</p>
        
        <div id="retry-area" style="margin-top: 30px; display: none; transition: opacity 0.5s;">
            <p style="font-size: 13px; color: #888; margin-bottom: 5px;">Proses memakan waktu lebih lama dari biasanya.</p>
            <span class="manual-link" onclick="window.location.href='${remoteURL}'">Muat Ulang Secara Manual</span>
        </div>
    </div>
`;

// Pengecekan Update Native
function doCheckUpdate() {
    CheckUpdate().then(result => {
        if (result.update_available) {
            // Panggil fungsi Go untuk memunculkan Popup Native OS
            ShowUpdatePrompt(result.latest_version, result.url);
        }
    }).catch(err => console.error("Update check failed:", err));
}

// Jalankan pengecekan update
doCheckUpdate();

// Redirect otomatis
setTimeout(() => {
    setTimeout(() => {
        const retry = document.getElementById('retry-area');
        if(retry) retry.style.display = 'block';
    }, 4000);
    window.location.assign(remoteURL);
}, 500);
