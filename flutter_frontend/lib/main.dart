import 'package:flutter/material.dart';
import 'package:flutter_frontend/router/app_router.dart';
import 'package:flutter_frontend/views/auth/auth_controller.dart';
import 'package:provider/provider.dart';

void main() {
  runApp(
    ChangeNotifierProvider(
      create: (context) => AuthController(),
      child: MyApp(), 
    ),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      routerConfig: goRouter,
      title: 'R-Time Demo',
      theme: ThemeData(
        // This is the theme of your application.
        //
        // TRY THIS: Try running your application with "flutter run". You'll see
        // the application has a purple toolbar. Then, without quitting the app,
        // try changing the seedColor in the colorScheme below to Colors.green
        // and then invoke "hot reload" (save your changes or press the "hot
        // reload" button in a Flutter-supported IDE, or press "r" if you used
        // the command line to start the app).
        //
        // Notice that the counter didn't reset back to zero; the application
        // state is not lost during the reload. To reset the state, use hot
        // restart instead.
        //
        // This works for code too, not just values: Most code changes can be
        // tested with just a hot reload.
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      // home: const MyHomePage(title: 'Flutter Demo Home Page'),
      // home: const LoginPage(),
      // home: InitAuthGate(),
    );
  }
}

// Deprecated; Router provides same service
// class InitAuthGate extends StatelessWidget {
//   const InitAuthGate({super.key});

//   @override
//   Widget build(BuildContext context) {
//     final authController = Provider.of<AuthController>(context);
//     return authController.isLoggedIn ? HomePage() : LoginPage();
//   }
// }