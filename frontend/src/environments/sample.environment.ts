// This file can be replaced during build by using the `fileReplacements` array.
// `ng build --prod` replaces `environment.ts` with `environment.prod.ts`.
// The list of file replacements can be found in `angular.json`.

export const environment = {
  production: false,
  firebaseConfig : {
    apiKey: "SET TO YOUR CREDS",
    authDomain: "SET TO YOUR CREDS",
    databaseURL: "SET TO YOUR CREDS",
    projectId: "SET TO YOUR CREDS",
    storageBucket: "SET TO YOUR CREDS",
    messagingSenderId: "SET TO YOUR CREDS",
    appId: "SET TO YOUR CREDS"
  },
  host_url: 'http://localhost:8080',
  client_id: "SET TO YOUR CREDS"
};

/*
 * For easier debugging in development mode, you can import the following file
 * to ignore zone related error stack frames such as `zone.run`, `zoneDelegate.invokeTask`.
 *
 * This import should be commented out in production mode because it will have a negative impact
 * on performance if an error is thrown.
 */
// import 'zone.js/dist/zone-error';  // Included with Angular CLI.
