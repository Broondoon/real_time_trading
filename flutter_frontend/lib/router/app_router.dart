// File initially generated via Gemini 2.0 Flash Experimental.
// Manually typed out and edited for my own understanding.
import 'package:flutter/material.dart';
import 'package:flutter_frontend/main.dart';
import 'package:flutter_frontend/views/auth/auth_controller.dart';
import 'package:flutter_frontend/views/auth/login_page_view.dart';
import 'package:flutter_frontend/views/home/home_page_view.dart';
import 'package:flutter_frontend/views/market/market_page_view.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

const String loginRouteName = 'login';
const String homeRouteName = 'home';
const String marketRouteName = 'market';

final goRouter = GoRouter(
  initialLocation: '/home',
  redirect: (context, state) {
    final authController = Provider.of<AuthController>(
      context,
      listen: false
    );
    // final isLoggedIn = true;
    final isLoggedIn = authController.isLoggedIn;
    final isLoggingIn = state.uri.toString() == '/login';

    // If we're not logged in, and not already at the login page, go to the login page.
    if (!isLoggedIn && !isLoggingIn) {
      print("Redirect to login.");
      return '/login';
    }

    // If we're logged in @ the login page, go to /home.
    if (isLoggedIn && isLoggingIn) {
      print("Redirect to home.");
      return '/home';
    }

    return null;
  },
  routes: [
    GoRoute(
      path: '/home',
      builder: (context, state) => const HomePage(),
      name: homeRouteName,
    ),
    GoRoute(
      path: '/login',
      builder: (context, state) => const LoginPage(),
      name: loginRouteName,
    ),
    GoRoute(
      path: '/market',
      builder: (context, state) => const MarketPage(),
      name: marketRouteName,
    ),
  ],
);