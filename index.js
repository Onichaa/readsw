const {
  default: WAConnect,
  useMultiFileAuthState,
  DisconnectReason,
  fetchLatestBaileysVersion,
  makeInMemoryStore,
  Browsers, 
  fetchLatestWaWebVersion
} = require("@whiskeysockets/baileys");
const pino = require("pino");
const readline = require('readline');
const { Boom } = require("@hapi/boom");

const pairingCode = process.argv.includes("--pairing-code");
const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
const question = (text) => new Promise((resolve) => rl.question(text, resolve));
const store = makeInMemoryStore({ logger: pino().child({ level: "silent", stream: "store" }) });


async function WAStart() {
  const { state, saveCreds } = await useMultiFileAuthState("./sesi");
  const { version, isLatest } = await fetchLatestWaWebVersion().catch(() => fetchLatestBaileysVersion());
  console.log(`using WA v${version.join(".")}, isLatest: ${isLatest}`);

  const client = WAConnect({
    logger: pino({ level: "silent" }),
    printQRInTerminal: !pairingCode,
    browser: Browsers.ubuntu("Chrome"),
    auth: state,
  });

  store.bind(client.ev);

  if (pairingCode && !client.authState.creds.registered) {
    const phoneNumber = await question(`Silahkan masukin nomor Whatsapp kamu: `);
    let code = await client.requestPairingCode(phoneNumber);
    code = code?.match(/.{1,4}/g)?.join("-") || code;
    console.log(`⚠︎ Kode Whatsapp kamu : ` + code)
  }

  client.ev.on("messages.upsert", async (chatUpdate) => {
    //console.log(JSON.stringify(chatUpdate, undefined, 2))
    try {
      const m = chatUpdate.messages[0];
      if (!m.message) return;
      if (m.key && !m.key.fromMe && m.key.remoteJid === 'status@broadcast') {
        await client.readMessages([m.key]);
        console.log("Berhasil melihat status", m.pushName)
      }
    } catch (err) {
      console.log(err);
    }
  });

  client.ev.on("connection.update", async (update) => {
    const { connection, lastDisconnect } = update;
      if (connection === "close") {
        let reason = new Boom(lastDisconnect?.error)?.output.statusCode;
        if (reason === DisconnectReason.badSession) {
          console.log(`Bad Session File, Please Delete Session and Scan Again`);
          process.exit();
        } else if (reason === DisconnectReason.connectionClosed) {
          console.log("Connection closed, reconnecting....");
          WAStart();
        } else if (reason === DisconnectReason.connectionLost) {
          console.log("Connection Lost from Server, reconnecting...");
          WAStart();
        } else if (reason === DisconnectReason.connectionReplaced) {
          console.log("Connection Replaced, Another New Session Opened, Please Restart Bot");
          process.exit();
        } else if (reason === DisconnectReason.loggedOut) {
          console.log(`Device Logged Out, Please Delete Folder Session and Scan Again.`);
          process.exit();
        } else if (reason === DisconnectReason.restartRequired) {
          console.log("Restart Required, Restarting...");
          WAStart();
        } else if (reason === DisconnectReason.timedOut) {
          console.log("Connection TimedOut, Reconnecting...");
          WAStart();
        } else {
          console.log(`Unknown DisconnectReason: ${reason}|${connection}`);
          WAStart();
        }
      } else if (connection === "open") {
      console.log("Connected to Readsw");
    }
  });

  client.ev.on("creds.update", saveCreds);

  return client;
}

WAStart();
