// vines-pos-desktop/frontend/src/main.js
import {TestPrint} from '../wailsjs/go/main/App';

const remoteURL = "http://45.64.97.50:888/thevines/index.php";

document.querySelector('#app').innerHTML = `
    <div style="display: flex; justify-content: center; align-items: center; height: 100vh; font-family: sans-serif; flex-direction: column; text-align: center; background-color: #f4f4f4;">
        <h2 id="status">Menghubungkan ke Vines POS...</h2>
        <p id="sub-status" style="color: gray;">Sedang mengalihkan ke server remote</p>
        
        <div id="retry-area" style="margin-top: 20px; display: none;">
            <p>Jika tidak kunjung terbuka, klik tombol di bawah:</p>
            <button onclick="window.location.href=remoteURL" style="padding: 10px 20px; background-color: #00CCFF; border: none; border-radius: 5px; color: white; cursor: pointer; font-weight: bold;">
                Buka Manual
            </button>
        </div>
    </div>
`;

window.remoteURL = remoteURL;

// Coba redirect otomatis setelah 1 detik
console.log("Attempting automatic redirect to: " + remoteURL);
setTimeout(() => {
    document.getElementById('retry-area').style.display = 'block';
    window.location.assign(remoteURL);
}, 1000);
