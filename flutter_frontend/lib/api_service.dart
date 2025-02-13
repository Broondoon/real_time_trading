import 'package:dio/dio.dart';
import 'package:flutter/material.dart';

class APIService {
  APIService(this.dio);
  final Dio dio;

  Future<Response?> post(context, String path, {dynamic data}) async {
    try {
      final response = await dio.post(
        path,
        data: data,
      );
      return response;
    }
    on DioException catch (e) {

      if (!context.mounted) {
        print(">>> ALERT! Could not display exception to use: $e");
        return null;
      }

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Request Failure: $e'),
        ),
      );

      return null;
    }
    catch (e) {
      print('>>> Something really bad happened. Generic exception: $e');
      rethrow;
    }
  }

  Future<Response?> get(context, String path) async {
    try {
      final response = await dio.get(path);
      return response;
    }
    on Exception catch (e) {
      if (!context.mounted) {
        print(">>> ALERT! Could not display exception to use: $e");
        return null;
      }

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Request Failure: $e'),
        ),
      );

      return null;
    }
  }
}

// TODO: in the future, see if standardizing a Response class that I control is useful.
// class BaseResponse<T> {
//   final int statusCode;
//   final String message;
//   final T? data;
//   // final bool success;

//   bool get isSuccessful => statusCode >= 200 && statusCode < 300;

//   BaseResponse({
//     required this.statusCode,
//     required this.message,
//     this.data,
//     // isSuccessful,
//   });

    // Pretend as if a factory .fromJSON is here.
// }