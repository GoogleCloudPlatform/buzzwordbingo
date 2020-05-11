export const environment = {
  production: true,
  firebaseConfig : {
    apiKey: "AIzaSyBau6rIjfndoZTdHRoIDAFSb0y8iC1HM_k",
    authDomain: "bingo-collab.firebaseapp.com",
    databaseURL: "https://bingo-collab.firebaseio.com",
    projectId: "bingo-collab",
    storageBucket: "bingo-collab.appspot.com",
    messagingSenderId: "1038359390820",
    appId: "1:1038359390820:web:392b2bf41e62ba61a14d71"
  },
  board_url: '/api/board',
  record_url: '/api/record',
  game_active_url: '/api/game/active',
  player_url: '/api/player/identify',
  admin_url: '/api/player/isadmin',
};
