// File initially generated via Gemini 2.0 Flash Experimental.
// Manually typed out and edited for my own understanding.
import 'package:flutter/material.dart';
import 'package:flutter_frontend/api_service.dart';
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
  late Dio _dio;
  // late APIService _apiService;

  // Lifetime should be 1 hour
  final int _tokenLifetime = 3600;
  DateTime? _tokenExpiration;

  AuthController() {
    _dio = Dio(
      BaseOptions(
        baseUrl: _baseUrl,
      )
    );

    // _apiService = APIService(
    //   _dio,
    // );

    // This is an INTERCEPTOR which adds JWT to Auth header
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          if (_token != null 
            && _tokenExpiration != null 
            && DateTime.now().isBefore(_tokenExpiration!)
          ) {
            options.headers['Authorization'] = 'Bearer $_token';
          }
          else {
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
          // // If the token has expired...
          // if (e.response?.statusCode == 401) {
          //   // Try for a new refreshed token.
          //   bool refreshed = await refreshToken();
            
          //   // And try again!
          //   if (refreshed) {
          //     return handler.resolve(
          //       await retry(e.requestOptions)
          //     );
          //   }
          //   else {
          //     // If that failed, we force move to login page.
          //     await logout();
          //     return handler.reject(e);
          //   }
          // }

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
      _dio.options.headers['Authorization'] = 'Bearer $_token';
    }
    else {
      _token = null;
      _tokenExpiration = null;
    }
    notifyListeners();
  }

  Future<bool> login(String username, String pwd) async {
    try {
      final response = await _dio.post(
        '/login', // TODO: replace this with a /resources/app_strings reference instead
        data: {
          'username': username,
          'password': pwd,
        },
      );

      // I now realize this is functionally USELESS until I create some unique
      //    behaviour that I control into the api service. TODO: do that
      // final response = await _apiService.post(
      //   '/login',
      //   data: {
      //     'username': username,
      //     'password': pwd,
      //   },
      // );
      // if (response == null) return false;

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
        _dio.options.headers['Authorization'] = 'Bearer $_token';
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

  Future<void> logout() async {
    _token = null;
    _tokenExpiration = null;
    await _storage.delete(
      key: 'jwt',
    );
    await _storage.delete(
      key: 'tokenExpiration',
    );
    _dio.options.headers.remove('Authorization');
    notifyListeners();
  }

  // This is just an example method
  Future<dynamic> fetchData() async {
    try {
      final response = await _dio.get('/protected');
      return response.data;
    }
    on DioException catch (e) {
      print('>> Exception: could not fetch data: $e');
      if (e.response?.statusCode == 401) {
        print('>> Token expired, which should be handled by the interceptor.');
      }
      rethrow;
    }
  }

}