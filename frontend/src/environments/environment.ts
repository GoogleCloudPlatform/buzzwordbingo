// This file can be replaced during build by using the `fileReplacements` array.
// `ng build --prod` replaces `environment.ts` with `environment.prod.ts`.
// The list of file replacements can be found in `angular.json`.

export const environment = {
  production: false,
  firebaseConfig : {
    apiKey: "AIzaSyBau6rIjfndoZTdHRoIDAFSb0y8iC1HM_k",
    authDomain: "bingo-collab.firebaseapp.com",
    databaseURL: "https://bingo-collab.firebaseio.com",
    projectId: "bingo-collab",
    storageBucket: "bingo-collab.appspot.com",
    messagingSenderId: "1038359390820",
    appId: "1:1038359390820:web:392b2bf41e62ba61a14d71"
  },
  board_url: 'http://127.0.0.1:8080/api/board',
};

/*
 * For easier debugging in development mode, you can import the following file
 * to ignore zone related error stack frames such as `zone.run`, `zoneDelegate.invokeTask`.
 *
 * This import should be commented out in production mode because it will have a negative impact
 * on performance if an error is thrown.
 */
// import 'zone.js/dist/zone-error';  // Included with Angular CLI.
