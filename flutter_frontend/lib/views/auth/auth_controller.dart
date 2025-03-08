// File initially generated via Gemini 2.0 Flash Experimental.
// Manually typed out and edited for my own understanding.
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:dio/dio.dart';

class AuthController extends ChangeNotifier {
  final _storage = const FlutterSecureStorage();
  String? _token;
  String? get token => _token; // A getter funct in one line
  
  // Login status getter: may need more logic in the future?
  bool get isLoggedIn => _token != null;

  // TODO: https
  final String _baseUrl = 'http://localhost:3001/';
  
  // Class objects; these could be dependency injected, no? Something to think about in the future.
  late Dio authdio;
  late Dio _unauthDio;
  // late APIService _apiService;

  // Lifetime should be 1 hour
  final int _tokenLifetime = 3600;
  DateTime? _tokenExpiration;

  AuthController() {
    authdio = Dio(
      BaseOptions(
        baseUrl: _baseUrl,
        // connectTimeout: const Duration(seconds: 10),
        // receiveTimeout: const Duration(seconds: 10),
      )
    );

    _unauthDio = Dio(
      BaseOptions(
        baseUrl: _baseUrl,
      )
    );

    // This is an INTERCEPTOR which adds JWT to Auth header
    authdio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          if (_token != null 
            && _tokenExpiration != null 
            && DateTime.now().isBefore(_tokenExpiration!)
          ) {
            options.headers['token'] = 'Bearer $_token';
          }
          else {            
            //TODO: re-enable auto-logout
            print("Logout redirect!");
            logout();
            return handler.reject(
              DioException(
                requestOptions: options,
                type: DioExceptionType.cancel,
              )
            );
          }
          return handler.next(options);
        },
        onError: (DioException e, handler) async {
          // ...let's just force logout if there's an issue.
          await logout();
          return handler.next(e);
        },
      ),
    );

    loadToken();
  }

  Future<void> loadToken() async {
    _token = await _storage.read(
      key: 'jwt',
    );
    final expirationString = await _storage.read(
      key: 'tokenExpiration'
    );

    if (_token != null && expirationString != null) {
      _tokenExpiration = DateTime.tryParse(expirationString);
      authdio.options.headers['token'] = 'Bearer $_token';
    }
    else {
      _token = null;
      _tokenExpiration = null;
    }
    notifyListeners();
  }

  Future<bool> login(String username, String pwd) async {
    try {
      print("Making login request:");

      var reqData = {
        'username': username,
        'password': pwd,
      };

      final response = await _unauthDio.post(
        '/authentication/login', // TODO: replace this with a /resources/app_strings reference instead
        data: jsonEncode(reqData),
      );

      print("STATUS CODE RESPONSE:");
      print(response.statusCode);

      if (response.statusCode == 200) {
        _token = response.data['token']; // TODO: double check what field the token is returned by
        _tokenExpiration = DateTime.now().add(
          Duration(seconds: _tokenLifetime)
        );
        await _storage.write(
          key: 'jwt',
          value: _token
        );
        await _storage.write(
          key: 'tokenExpiration',
          value: _tokenExpiration!.toIso8601String()
        );
        authdio.options.headers['token'] = 'Bearer $_token';
        notifyListeners();
        return true;
      }
      else {
        // TODO: Yeah, a logging framework would be nice to include!
        print('>> Login failed: ${response.statusCode}');
        return false;
      }
    }
    on DioException catch (e) {
      print('>> Exception thrown during login: $e');
      return false;
    }
  }

    Future<bool> register(String username, String pwd) async {
    try {
      print("Making registration request:");
      final response = await _unauthDio.post(
        '/authentication/register', // TODO: replace this with a /resources/app_strings reference instead
        data: {
          'username': username,
          'password': pwd,
          // 'name': 'Test User', //TODO: Auth isn't prepared to accept the name yet
        },
      );

      print("STATUS CODE RESPONSE:");
      print(response.statusCode);

      if (response.statusCode == 200) {
        print("Nice! Registered.");
        return true;
      }
      else {
        // TODO: Yeah, a logging framework would be nice to include!
        print('>> Registration failed: ${response.statusCode}');
        return false;
      }
    }
    on DioException catch (e) {
      print('>> Exception thrown during registration: $e');
      return false;
    }
  }

  Future<void> logout() async {
    _token = null;
    _tokenExpiration = null;
    await _storage.delete(
      key: 'jwt',
    );
    await _storage.delete(
      key: 'tokenExpiration',
    );
    authdio.options.headers.remove('token');
    notifyListeners();
  }
}
